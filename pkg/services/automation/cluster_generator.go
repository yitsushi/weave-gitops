package automation

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"os"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/git"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"

	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"

	"github.com/weaveworks/weave-gitops/pkg/models"
	"sigs.k8s.io/yaml"
)

type ClusterAutomation struct {
	AppCRD                      AutomationManifest
	GitOpsRuntime               AutomationManifest
	SourceManifest              AutomationManifest
	SystemKustomizationManifest AutomationManifest
	SystemKustResourceManifest  AutomationManifest
	UserKustResourceManifest    AutomationManifest
	WegoAppManifest             AutomationManifest
}

const (
	AppCRDPath              = "wego-system.yaml"
	RuntimePath             = "gitops-runtime.yaml"
	SourcePath              = "flux-source-resource.yaml"
	SystemKustomizationPath = "kustomization.yaml"
	SystemKustResourcePath  = "flux-system-kustomization-resource.yaml"
	UserKustResourcePath    = "flux-user-kustomization-resource.yaml"
	WegoAppPath             = "wego-app.yaml"
	WegoConfigPath          = "wego-config.yaml"

	WegoConfigMapName = "weave-gitops-config"
)

func createClusterSourceName(gitSourceURL gitproviders.RepoURL) string {
	provider := string(gitSourceURL.Provider())
	cleanRepoName := replaceUnderscores(gitSourceURL.RepositoryName())
	qualifiedName := fmt.Sprintf("wego-auto-%s-%s", provider, cleanRepoName)
	lengthConstrainedName := hashNameIfTooLong(qualifiedName)

	return lengthConstrainedName
}

type ClusterAutomationParams struct {
	ConfigURL      gitproviders.RepoURL
	Branch         string
	RepoVisibility gitprovider.RepositoryVisibility

	Cluster models.Cluster

	CreateNamespace bool
	Namespace       string
}

func (a *AutomationGen) GenerateClusterAutomation(ctx context.Context, params ClusterAutomationParams) (ClusterAutomation, error) {
	secretRef, err := a.GetSecretRefForPrivateGitSources(ctx, params.ConfigURL)
	if err != nil {
		return ClusterAutomation{}, err
	}

	secretStr := secretRef.String()

	configBranch, err := a.GitProvider.GetDefaultBranch(ctx, params.ConfigURL)
	if err != nil {
		return ClusterAutomation{}, err
	}

	runtimeManifests, err := a.Flux.Install(params.Namespace, true)
	if err != nil {
		return ClusterAutomation{}, err
	}

	appCRDManifest := manifests.AppCRD

	version := version.Version
	if os.Getenv("IS_TEST_ENV") != "" {
		version = "latest"
	}

	m, err := manifests.GenerateManifests(manifests.Params{AppVersion: version, Namespace: params.Namespace, CreateNamespace: params.CreateNamespace})
	if err != nil {
		return ClusterAutomation{}, fmt.Errorf("error generating wego-app manifest: %w", err)
	}

	wegoAppManifest := bytes.Join(m, []byte("---\n"))

	sourceName := createClusterSourceName(params.ConfigURL)

	sourceManifest, err := a.Flux.CreateSourceGit(sourceName, params.ConfigURL, configBranch, secretStr, params.Namespace)
	if err != nil {
		return ClusterAutomation{}, err
	}

	systemKustResourceManifest, err := a.Flux.CreateKustomization(ConstrainResourceName(fmt.Sprintf("%s-system", params.Cluster.Name)), sourceName,
		workAroundFluxDroppingDot(git.GetSystemPath(params.Cluster.Name)), params.Namespace)
	if err != nil {
		return ClusterAutomation{}, err
	}

	userKustResourceManifest, err := a.Flux.CreateKustomization(ConstrainResourceName(fmt.Sprintf("%s-user", params.Cluster.Name)), sourceName,
		workAroundFluxDroppingDot(git.GetUserPath(params.Cluster.Name)), params.Namespace)
	if err != nil {
		return ClusterAutomation{}, err
	}

	systemKustomization := CreateKustomize(params.Cluster.Name, params.Namespace, RuntimePath, SourcePath, SystemKustResourcePath, UserKustResourcePath)

	systemKustomizationManifest, err := yaml.Marshal(systemKustomization)
	if err != nil {
		return ClusterAutomation{}, err
	}

	return ClusterAutomation{
		AppCRD: AutomationManifest{
			Path:    git.GetSystemQualifiedPath(params.Cluster.Name, AppCRDPath),
			Content: appCRDManifest,
		},
		GitOpsRuntime: AutomationManifest{
			Path:    git.GetSystemQualifiedPath(params.Cluster.Name, RuntimePath),
			Content: runtimeManifests,
		},
		SourceManifest: AutomationManifest{
			Path:    git.GetSystemQualifiedPath(params.Cluster.Name, SourcePath),
			Content: sourceManifest,
		},
		SystemKustomizationManifest: AutomationManifest{
			Path:    git.GetSystemQualifiedPath(params.Cluster.Name, SystemKustomizationPath),
			Content: systemKustomizationManifest,
		},
		SystemKustResourceManifest: AutomationManifest{
			Path:    git.GetSystemQualifiedPath(params.Cluster.Name, SystemKustResourcePath),
			Content: systemKustResourceManifest,
		},
		UserKustResourceManifest: AutomationManifest{
			Path:    git.GetSystemQualifiedPath(params.Cluster.Name, UserKustResourcePath),
			Content: userKustResourceManifest,
		},
		WegoAppManifest: AutomationManifest{
			Path:    git.GetSystemQualifiedPath(params.Cluster.Name, WegoAppPath),
			Content: wegoAppManifest,
		},
	}, nil
}

func (ca ClusterAutomation) BootstrapManifests() []AutomationManifest {
	return append([]AutomationManifest{ca.AppCRD}, ca.WegoAppManifest, ca.SourceManifest, ca.SystemKustResourceManifest, ca.UserKustResourceManifest)
}

func (ca ClusterAutomation) GenerateWegoConfigManifest(clusterName string, fluxNamespace string, wegoNamespace string) (AutomationManifest, error) {
	config := kube.WegoConfig{
		FluxNamespace: fluxNamespace,
		WegoNamespace: wegoNamespace,
	}

	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return AutomationManifest{}, fmt.Errorf("failed marshalling wego config: %w", err)
	}

	name := types.NamespacedName{Name: WegoConfigMapName, Namespace: wegoNamespace}

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Data: map[string]string{
			"config": string(configBytes),
		},
	}

	wegoConfigManifest, err := yaml.Marshal(cm)
	if err != nil {
		return AutomationManifest{}, fmt.Errorf("failed marshalling wego config: %w", err)
	}

	return AutomationManifest{
		Path:    git.GetSystemQualifiedPath(clusterName, WegoConfigPath),
		Content: wegoConfigManifest,
	}, nil
}

func (ca ClusterAutomation) Manifests() []AutomationManifest {
	return append(ca.BootstrapManifests(), ca.GitOpsRuntime, ca.SystemKustomizationManifest)
}

func GetClusterHash(c models.Cluster) string {
	return fmt.Sprintf("wego-%x", md5.Sum([]byte(c.Name)))
}

func workAroundFluxDroppingDot(str string) string {
	return "." + str
}
