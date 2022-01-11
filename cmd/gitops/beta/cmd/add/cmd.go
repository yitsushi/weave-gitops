package add

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/beta/cmd/add/app"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new Weave GitOps resource",
		Example: `
# Add an application to gitops from local git repository
gitops add app . --name <app-name>`,
	}

	cmd.AddCommand(app.AppCmd)

	return cmd
}
