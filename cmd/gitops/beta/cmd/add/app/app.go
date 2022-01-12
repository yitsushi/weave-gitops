/*
Copyright © 2021 Weaveworks <support@weave.works>
This file is part of the Weave GitOps CLI.
*/
package app

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/core/gitops/app"
	"github.com/weaveworks/weave-gitops/core/repository"
)

const (
	namespaceFlag   = "namespace"
	descriptionFlag = "description"
	displayNameFlag = "display-name"
)

type Params struct {
	Name        string
	Namespace   string
	Description string
	DisplayName string
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
	AppCmd.Flags().StringVar(&params.Namespace, namespaceFlag, "", "Namespace for the app")
	AppCmd.Flags().StringVar(&params.Description, descriptionFlag, "", "Description of the app")
	AppCmd.Flags().StringVar(&params.Namespace, displayNameFlag, "", "Display name to use in the Weave GitOps app")
}

func runCmd(cmd *cobra.Command, args []string) error {
	r := bufio.NewReader(os.Stdin)
	return createApp(r)
}

func createApp(r *bufio.Reader) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to determine local directory: %w", err)
	}

	repo, err := internal.GitRepository(dir)
	if err != nil {
		return err
	}

	gitCommitter := repository.NewGitWriter(false)
	appSvc := app.NewCreator(gitCommitter)

	if params.DisplayName == "" {
		fmt.Printf("Display Name: ")

		params.Namespace, err = readAndFormatInput(r, "display name")
		if err != nil {
			return err
		}
	}

	if params.Namespace == "" {
		fmt.Printf("Namespace (e.g. flux-system): ")

		params.Namespace, err = readAndFormatInput(r, "namespace")
		if err != nil {
			return err
		}
	}

	if params.Description == "" {
		fmt.Printf("Description: ")

		params.Description, err = readAndFormatInput(r, "description")
		if err != nil {
			return err
		}
	}

	_, err = appSvc.Create(repo, internal.StubAuth{}, app.CreateInput{
		Name:        params.Name,
		Namespace:   params.Namespace,
		Description: params.Description,
	})
	if err != nil {
		return fmt.Errorf("issue creating an app: %w", err)
	}

	return nil
}

func readAndFormatInput(r *bufio.Reader, field string) (string, error) {
	input, err := r.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("issue reading input for %s: %w", field, err)
	}

	input = strings.Replace(input, "\n", "", -1)
	return input, nil
}
