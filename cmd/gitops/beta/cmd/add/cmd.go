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
# Add a Flux Kustomization to an app from local git repository
gitops add kustomization <kust-name> --app <app-name>`,
	}

	cmd.AddCommand(app.KustomizationCmd)

	return cmd
}
