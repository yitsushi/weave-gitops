/*
Copyright © 2021 Weaveworks <support@weave.works>
This file is part of the Weave GitOps CLI.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/core/gitops/install"
	"github.com/weaveworks/weave-gitops/core/repository"
)

type params struct {
	FluxPath []string
}

var (
	installParams params
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install or upgrade GitOps",
	Long: `The beta install command creates the manifests in the current directory.  The directory must
be an initialized git repository and have Flux installed.`,
	Example: `  # Install GitOps in the wego-system namespace
  gitops beta install --config-repo ssh://git@github.com/me/mygitopsrepo.git`,
	RunE:          installRunCmd,
	SilenceErrors: true,
	SilenceUsage:  true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

func init() {
	Cmd.AddCommand(installCmd)
	installCmd.Flags().StringSliceVar(&installParams.FluxPath, "flux-paths", []string{}, "List of flux's gitops toolkit paths to install Weave GitOps.  E.g. ./dev-cluster/flux-system,./staging-cluster/flux-system")
	//cobra.CheckErr(installCmd.MarkFlagRequired("flux-paths"))
}

func installRunCmd(_ *cobra.Command, _ []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to determine local directory: %w", err)
	}

	repo, err := internal.GitRepository(dir)
	if err != nil {
		return err
	}

	gitopsInstaller := install.NewGitopsInstaller(repository.NewGitCommitter())

	err = gitopsInstaller.Install(repo)
	if err != nil {
		fmt.Errorf("there was an issue installing Weave Gitops: %w", err)
	}

	return nil
}
