package get

import (
	"github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a Weave GitOps resource",
		Example: `
# Get an application to gitops from local git repository
gitops get app <app-name>`,
	}

	cmd.AddCommand(AppCmd)

	return cmd
}
