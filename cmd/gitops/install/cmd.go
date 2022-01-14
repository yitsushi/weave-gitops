package install

// Provides support for adding a repository of manifests to a gitops cluster. If the cluster does not have
// gitops installed, the user will be prompted to install gitops and then the repository will be added.

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/pkg/services/applier"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"github.com/weaveworks/weave-gitops/pkg/services/gitopswriter"
	"github.com/weaveworks/weave-gitops/pkg/services/gitrepo"
)

type params struct {
	DryRun          bool
	AutoMerge       bool
	ConfigRepo      string
	SkipFluxInstall bool
	Namespace       string
}

var (
	installParams params
)

var Cmd = &cobra.Command{
	Use:   "install",
	Short: "Install or upgrade GitOps",
	Long: `The install command deploys GitOps in the specified namespace,
adds a cluster entry to the GitOps repo, and persists the GitOps runtime into the
repo. If a previous version is installed, then an in-place upgrade will be performed.`,
	Example: fmt.Sprintf(`  # Install GitOps in the %s namespace
  gitops install --config-repo=ssh://git@github.com/me/mygitopsrepo.git`, wego.DefaultNamespace),
	RunE:          installRunCmd,
	SilenceErrors: true,
	SilenceUsage:  true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

const LabelPartOf = "app.kubernetes.io/part-of"

func init() {
	Cmd.Flags().BoolVar(&installParams.DryRun, "dry-run", false, "Outputs all the manifests that would be installed")
	Cmd.Flags().BoolVar(&installParams.AutoMerge, "auto-merge", false, "If set, 'gitops install' will automatically update the default branch for the configuration repository")
	Cmd.Flags().StringVar(&installParams.ConfigRepo, "config-repo", "", "URL of external repository that will hold automation manifests")
	Cmd.Flags().BoolVar(&installParams.SkipFluxInstall, "skip-flux-install", false, "Skips Flux installation")
	cobra.CheckErr(Cmd.MarkFlagRequired("config-repo"))
}

func installRunCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	installParams.Namespace, _ = cmd.Parent().Flags().GetString("namespace")

	configURL, err := gitproviders.NewRepoURL(installParams.ConfigRepo)
	if err != nil {
		return err
	}

	osysClient := osys.New()
	log := internal.NewCLILogger(os.Stdout)
	flux := flux.New(osysClient, &runner.CLIRunner{})

	k, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("error creating k8s http client: %w", err)
	}

	if err := validateWegoInstall(ctx, k, installParams); err != nil {
		return err
	}

	clusterName, err := k.GetClusterName(ctx)
	if err != nil {
		return err
	}

	clusterApplier := applier.NewClusterApplier(k)

	var gitClient git.Git

	factory := services.NewFactory(flux, log)

	if !installParams.SkipFluxInstall {
		_, err = flux.Install(installParams.Namespace, false)
		if err != nil {
			return err
		}
	}

	providerClient := internal.NewGitProviderClient(osysClient.Stdout(), osysClient.LookupEnv, auth.NewAuthCLIHandler, log)

	gitProvider, err := providerClient.GetProvider(configURL, gitproviders.GetAccountType)
	if err != nil {
		return fmt.Errorf("error obtaining git provider token: %w", err)
	}

	repoVisibility, err := gitProvider.GetRepoVisibility(ctx, configURL)
	if err != nil {
		return fmt.Errorf("failed getting config repo visibility: %w", err)
	}

	repoBranch, err := gitProvider.GetDefaultBranch(ctx, configURL)
	if err != nil {
		return fmt.Errorf("failed getting default branch for config repo: %w", err)
	}

	cluster := models.Cluster{Name: clusterName}

	automationGen := automation.NewAutomationGenerator(gitProvider, flux, log)

	clusterAutomation, err := automationGen.GenerateClusterAutomation(ctx, automation.ClusterAutomationParams{
		Cluster:         cluster,
		ConfigURL:       configURL,
		Namespace:       installParams.Namespace,
		RepoVisibility:  *repoVisibility,
		Branch:          repoBranch,
		CreateNamespace: installParams.SkipFluxInstall,
	})
	if err != nil {
		return err
	}

	wegoConfigManifest, err := clusterAutomation.GenerateWegoConfigManifest(clusterName, installParams.Namespace, installParams.Namespace)
	if err != nil {
		return fmt.Errorf("failed generating wego config manifest: %w", err)
	}

	manifests := append(clusterAutomation.Manifests(), wegoConfigManifest)

	if installParams.DryRun {
		for _, manifest := range manifests {
			log.Println(string(manifest.Content))
		}

		return nil
	}

	err = clusterApplier.ApplyManifests(ctx, cluster, installParams.Namespace, append(clusterAutomation.BootstrapManifests(), wegoConfigManifest))
	if err != nil {
		return fmt.Errorf("failed applying manifest: %w", err)
	}

	gitClient, _, err = factory.GetGitClients(context.Background(), providerClient, services.GitConfigParams{
		URL:       installParams.ConfigRepo,
		Namespace: installParams.Namespace,
		DryRun:    installParams.DryRun,
	})
	if err != nil {
		return fmt.Errorf("error creating git clients: %w", err)
	}

	repoWriter := gitrepo.NewRepoWriter(configURL, gitProvider, gitClient, log)
	gitOpsDirWriter := gitopswriter.NewGitOpsDirectoryWriter(automationGen, repoWriter, osysClient, log)

	err = gitOpsDirWriter.AssociateCluster(ctx, cluster, configURL, installParams.Namespace, installParams.Namespace, installParams.AutoMerge, installParams.SkipFluxInstall, repoBranch, manifests)
	if err != nil {
		return fmt.Errorf("failed associating cluster: %w", err)
	}

	return nil
}

func validateWegoInstall(ctx context.Context, kubeClient kube.Kube, params params) error {
	fluxPresent, err := kubeClient.FluxPresent(ctx)
	if err != nil {
		return fmt.Errorf("failed checking flux presence: %w", err)
	}

	if fluxPresent {
		if !params.SkipFluxInstall {
			return errors.New("There is a standalone installation of Flux in your cluster.\n\nTo avoid conflict you should either:\nSkip Flux installation:\n  $ gitops install --config-repo [config-repo] --skip-flux-install\n\nOr uninstall flux before proceeding:\n  $ flux uninstall")
		}
	}

	status := kubeClient.GetClusterStatus(ctx)
	if status == kube.Unknown {
		return errors.New("Weave GitOps cannot talk to the cluster")
	}

	wegoConfig, err := kubeClient.GetWegoConfig(ctx, "")
	if err != nil {
		if !errors.Is(err, kube.ErrWegoConfigNotFound) {
			return fmt.Errorf("Failed getting wego config: %w", err)
		}
	}

	if wegoConfig.WegoNamespace != "" && wegoConfig.WegoNamespace != params.Namespace {
		return errors.New("You cannot install Weave GitOps into a different namespace")
	}

	return nil
}
