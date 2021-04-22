package install

// Provides support for adding a repository of manifests to a wego cluster. If the cluster does not have
// wego installed, the user will be prompted to install wego and then the repository will be added.

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

//go:embed manifests/bootstrap/*.yaml
var embeddedManifests embed.FS

//go:embed manifests/kpack/builder.yaml
var embeddedBuilderManifest []byte

type paramSet struct {
	owner      string
	repository string
	branch     string
	path       string
}

var (
	params    paramSet
	repoOwner string
)

var Cmd = &cobra.Command{
	Use:   "install [--owner <owner>] [--repository <repository>] [--branch <branch>] [--path <path>]",
	Short: "Install wego components into the k8s cluster",
	Long: strings.TrimSpace(dedent.Dedent(`
    `)),
	Example: "wego install",
	Run:     runCmd,
}

// checkError will print a message to stderr and exit
func checkError(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		os.Exit(1)
	}
}

func init() {
	Cmd.Flags().StringVar(&params.owner, "owner", "", "Name of remote git repository")
	Cmd.Flags().StringVar(&params.repository, "repository", "", "URL of remote git repository")
	Cmd.Flags().StringVar(&params.branch, "branch", "main", "Branch to watch within git repository")
	Cmd.Flags().StringVar(&params.path, "path", "./clusters/my-cluster", "Path to watch within git repository")
}

func runCmd2(cmd *cobra.Command, args []string) {
	utils.CallCommand(fmt.Sprintf("%s %s", "kubectl", strings.Join([]string{"get", "pods", "-A"}, " ")))
}

func runCmd(cmd *cobra.Command, args []string) {
	fluxArgs := []string{
		"--personal=true",
		"--components-extra=image-reflector-controller,image-automation-controller",
		"--read-write-key=true",
		fmt.Sprintf("--owner=%s", params.owner),
		fmt.Sprintf("--repository=%s", params.repository),
		fmt.Sprintf("--branch=%s", params.branch),
		fmt.Sprintf("--path=%s", params.path),
	}
	_, err := fluxops.CallFlux(fmt.Sprintf("bootstrap github %s", strings.Join(fluxArgs, " ")))
	checkError("failed to run flux", err)

	err = git.PushChangesChangesToWegoRepo(params.owner, params.repository, func(repoDir string) string {
		clusterDir := path.Join(repoDir, "clusters", "my-cluster")

		// Writing kpack components to wego repo
		err := writeKpackEmbeddedManifests(clusterDir)
		checkError("failed to generate kpack manifests", err)

		// Writting builder file to kpack directory
		bootstrapDir := path.Join(clusterDir, "kpack")
		os.MkdirAll(bootstrapDir, os.ModePerm)

		err = os.WriteFile(path.Join(clusterDir, "kpack", "builder.yaml"), embeddedBuilderManifest, 0666)
		checkError("failed to writing kpack builder manifest", err)

		// Applying kpack components to wego repo. This is a necessary because of the crds.
		_, err = utils.CallCommand(fmt.Sprintf("kubectl %s", strings.Join([]string{"apply", "-k", path.Join(clusterDir, "bootstrap")}, " ")))
		checkError("failed to apply kpack manifests", err)

		return "Add kpack manifests"
	})
	checkError("failed to push changes to github", err)
}

func writeKpackEmbeddedManifests(clusterDir string) error {
	embeddedDir := "manifests/bootstrap"
	manifests, err := fs.ReadDir(embeddedManifests, embeddedDir)
	if err != nil {
		return err
	}
	for _, manifest := range manifests {
		if manifest.IsDir() {
			continue
		}

		data, err := fs.ReadFile(embeddedManifests, path.Join(embeddedDir, manifest.Name()))
		if err != nil {
			return fmt.Errorf("reading file failed: %w", err)
		}

		bootstrapDir := path.Join(clusterDir, "bootstrap")
		os.MkdirAll(bootstrapDir, os.ModePerm)

		err = os.WriteFile(path.Join(bootstrapDir, manifest.Name()), data, 0666)
		if err != nil {
			return fmt.Errorf("writing file failed: %w", err)
		}
	}
	return nil
}
