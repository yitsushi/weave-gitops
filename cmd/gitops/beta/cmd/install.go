/*
Copyright © 2021 Weaveworks <support@weave.works>
This file is part of the Weave GitOps CLI.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/git"
)

type params struct {
	DryRun     bool
	ConfigRepo string
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
	installCmd.Flags().BoolVar(&installParams.DryRun, "dry-run", false, "Outputs all the manifests that would be installed")
	installCmd.Flags().StringVar(&installParams.ConfigRepo, "config-repo", "", "URL of external repository that will hold automation manifests")
	//cobra.CheckErr(installCmd.MarkFlagRequired("config-repo"))
}

func installRunCmd(cmd *cobra.Command, args []string) error {
	branch, err := internal.LocalBranch()
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", branch)

	config :=

		git.Commit{
			Author:  git.Author{},
			Hash:    "",
			Message: "",
		}

	return nil
}
