/*
Copyright © 2021 Weaveworks <support@weave.works>
This file is part of the Weave GitOps CLI.
*/
package get

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/core/gitops/app"
)

const (
	nameFlag = "name"
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
	Short: "Get an application from the GitOps repository",
	Long: `This command will get an application from the directory structure. It is
	required that the app is located within \".weave-gitops/apps/<app-name\".`,
	RunE: runCmd,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			params.Name = args[0]
		}

		return nil
	},
}

func init() {
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

	appSvc := app.NewRepoFetcher()

	if params.Name == "" {
		apps, err := appSvc.List(dir)
		if err != nil {
			return fmt.Errorf("unable to get list of apps: %w", err)
		}

		for _, app := range apps {
			fmt.Println(app.Name)
		}
	} else {
		if app, err := appSvc.Get(dir, params.Name); err != nil {
			return fmt.Errorf("unable to get app: %w", err)
		} else {
			data, err := json.MarshalIndent(app, "", "  ")
			if err != nil {
				return fmt.Errorf("unable to transform app into json: %w", err)
			}

			fmt.Printf("%s", data)
		}
	}

	return nil
}
