/*
Copyright © 2021 Weaveworks <support@weave.works>
This file is part of the Weave GitOps CLI.
*/
package app

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/core/gitops/app"
	"github.com/weaveworks/weave-gitops/core/repository"
)

type Params struct {
	Name string
}

var (
	params Params
)

// appCmd represents the app command
var AppCmd = &cobra.Command{
	Use:   "app",
	Short: "Adds an application workload to the GitOps repository",
	Long: `This command mirrors the original add app command in
	that it adds the definition for the application to the repository
	and sets up syncing into a cluster. It uses the new directory
	structure.`,
	RunE: runCmd,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("add app requires a name argument")
		}
		params.Name = args[0]
		return nil
	},
}

func init() {

}

func runCmd(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to determine local directory: %w", err)
	}

	repo, err := internal.GitRepository(dir)
	if err != nil {
		return err
	}

	gitCommitter := repository.NewGitCommitter()
	appSvc := app.NewCreator(gitCommitter)

	_, err = appSvc.Create(repo, params.Name, "test-space", "This is a test")
	if err != nil {
		return fmt.Errorf("issue creating an app: %w", err)
	}

	return nil
}
