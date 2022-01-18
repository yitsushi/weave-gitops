package create

import (
	"github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new Weave GitOps resource",
		Example: `
# Get an application to gitops from local git repository
gitops get app <app-name>`,
	}

	cmd.AddCommand(AppCmd)

	return cmd
}
