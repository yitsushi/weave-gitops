package add

// Provides support for adding a repository of manifests to a wego cluster. If the cluster does not have
// wego installed, the user will be prompted to install wego and then the repository will be added.

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/k0kubun/pp"
	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/status"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

//go:embed manifests/bootstrap/kpack-bar.yaml
var embeddedManifests embed.FS

type paramSet struct {
	name   string
	url    string
	branch string
}

var (
	params    paramSet
	repoOwner string
)

var Cmd = &cobra.Command{
	Use:   "add [--name <name>] [--url <url>] [--branch <branch>] <repository directory>",
	Short: "Add a workload repository to a wego cluster",
	Long: strings.TrimSpace(dedent.Dedent(`
        Associates an additional git repository with a wego cluster so that its contents may be managed via GitOps
    `)),
	Example: "wego add",
	Run:     runCmd2,
}

func runCmd2(cmd *cobra.Command, args []string) {
	// // creating temp dir to dump kpack manifests to copy to wego repo
	// tmpDir, err := ioutil.TempDir("", "wego-bootstrap-")
	// checkAddError(err)

	// // defer os.RemoveAll(tmpDir)
	// fmt.Println("Saving files to: ", tmpDir)
	// err = writeEmbeddedManifests(tmpDir)
	// checkAddError(err)

	cloneRepo()
}

func cloneRepo() {
	tmpDir, err := ioutil.TempDir("", "wego-repo-")
	// defer os.RemoveAll(tmpDir)
	checkAddError(err)
	pp.Println(tmpDir)

	publicKeys, err := ssh.NewPublicKeysFromFile("git", filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"), "")
	checkAddError(err)

	repository := "git@github.com:luizbafilho/fleet-infra.git"
	r, err := git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:  repository,
		Auth: publicKeys,
	})
	checkAddError(err)
	fmt.Println("Repository cloned")

	w, err := r.Worktree()
	checkAddError(err)

	manifests, err := w.Filesystem.ReadDir(".")
	checkAddError(err)
	for _, manifest := range manifests {
		pp.Println(manifest.Name())
	}

	err = writeKpackEmbeddedManifests(filepath.Join(tmpDir, "clusters", "my-cluster"))
	checkAddError(err)

	err = w.AddGlob(".")
	checkAddError(errors.Wrap(err, "failed to add"))

	_, err = w.Commit("add kpack controller and service account", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})
	checkAddError(errors.Wrap(err, "failed to commit"))

	err = r.Push(&git.PushOptions{
		Auth: publicKeys,
	})
	checkAddError(err)
}

func writeKpackEmbeddedManifests(dir string) error {
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

		kpackDir := dir + "/kpack"
		os.MkdirAll(kpackDir, os.ModePerm)

		err = os.WriteFile(path.Join(kpackDir, manifest.Name()), data, 0666)
		if err != nil {
			return fmt.Errorf("writing file failed: %w", err)
		}
	}
	return nil
}

// checkError will print a message to stderr and exit
func checkError(msg string, err interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		os.Exit(1)
	}
}

func checkAddError(err interface{}) {
	checkError("Failed to add workload repository", err)
}

func init() {
	Cmd.Flags().StringVar(&params.name, "name", "", "Name of remote git repository")
	Cmd.Flags().StringVar(&params.url, "url", "", "URL of remote git repository")
	Cmd.Flags().StringVar(&params.branch, "branch", "main", "Branch to watch within git repository")
}

func updateParametersIfNecessary() {
	if params.url == "" {
		urlout, err := exec.Command("git", "remote", "get-url", "origin").CombinedOutput()
		checkError("Failed to discover URL of remote repository", err)
		url := strings.TrimRight(string(urlout), "\n")
		fmt.Printf("URL not specified; ")
		params.url = url
	}

	sshPrefix := "git@github.com:"
	if strings.HasPrefix(params.url, sshPrefix) {
		repoName, err := fluxops.GetRepoName()
		checkAddError(err)
		isPrivate, err := fluxops.IsPrivate(getOwner(), repoName)
		if err != nil {
			isPrivate = askUser("Should the WeGO repository be private? (Y/n)")
		}
		if isPrivate {
			params.url = "ssh://git@github.com/" + strings.TrimPrefix(params.url, sshPrefix)
		} else {
			params.url = "https://github.com/" + strings.TrimPrefix(params.url, sshPrefix)
		}
	}

	fmt.Printf("using URL: '%s' of origin from git config...\n\n", params.url)

	if params.name == "" {
		clusterName, err := status.GetClusterName()
		checkAddError(err)
		params.name = clusterName + "-wego"
	}
}

func generateSourceManifest(repoName string) []byte {
	sourceManifest, err := fluxops.CallFlux(fmt.Sprintf(`create source git "%s" --url="%s" --branch="%s" --interval=30s --export`,
		repoName, params.url, params.branch))
	checkAddError(err)
	return sourceManifest
}

func generateKustomizeManifest(repoName string) []byte {
	kustomizeManifest, err := fluxops.CallFlux(
		fmt.Sprintf(`create kustomization "%s" --path="./" --source="%s" --prune=true --validation=client --interval=5m --export`, params.name, repoName))
	checkAddError(err)
	return kustomizeManifest
}

func bootstrapOrExit() {
	if !askUser("The cluster does not have wego installed; install it now? (Y/n)") {
		fmt.Fprintf(os.Stderr, "Wego not installed.")
		os.Exit(1)
	}
	repoName, err := fluxops.GetRepoName()
	checkAddError(err)
	fluxops.Bootstrap(getOwner(), repoName)

}

func askUser(question string) bool {
	fmt.Printf("%s ", question)
	return proceed()
}

func proceed() bool {
	answer := getAnswer()
	for !validAnswer(answer) {
		fmt.Println("Invalid answer, please choose 'Y' or 'n'")
		answer = getAnswer()
	}
	return strings.EqualFold(answer, "y")
}

func getAnswer() string {
	reader := bufio.NewReader(os.Stdin)
	str, err := reader.ReadString('\n')
	checkAddError(err)
	if str == "\n" {
		str = "Y\n"
	}
	return strings.Trim(str, "\n")
}

func validAnswer(answer string) bool {
	return strings.EqualFold(answer, "y") || strings.EqualFold(answer, "n")
}

func getOwner() string {
	if repoOwner != "" {
		return repoOwner
	}
	owner, err := fluxops.GetOwnerFromEnv()
	if err != nil || owner == "" {
		repoOwner = getOwnerInteractively()
		return repoOwner
	}
	repoOwner = owner
	return owner
}

func getOwnerInteractively() string {
	fmt.Printf("Who is the owner of the repository? ")
	reader := bufio.NewReader(os.Stdin)
	str, err := reader.ReadString('\n')
	checkAddError(err)

	if str == "\n" {
		return getOwnerInteractively()
	}

	return strings.Trim(str, "\n")
}

func commitAndPush(files ...string) {
	_, err := utils.CallCommand(
		fmt.Sprintf("git pull --rebase && git add %s && git commit -m'Save %s' && git push", strings.Join(files, " "), strings.Join(files, ", ")))
	checkAddError(err)
}

func runCmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Printf("Location of application not specified.\n")
		os.Exit(1)
	}
	fmt.Printf("Updating parameters from environment... done\n\n")
	updateParametersIfNecessary()
	fmt.Printf("Checking cluster status... ")
	clusterStatus := status.GetClusterStatus()
	fmt.Printf("%s\n\n", clusterStatus)
	if clusterStatus == status.Unmodified {
		bootstrapOrExit()
	}

	repoPath, err := filepath.Abs(args[0])
	checkAddError(err)
	repoName := filepath.Base(repoPath)
	source := generateSourceManifest(repoName)
	kust := generateKustomizeManifest(repoName)

	fluxRepoName, err := fluxops.GetRepoName()
	checkAddError(err)

	fluxRepo := filepath.Join(os.Getenv("HOME"), ".wego", "repositories", fluxRepoName)
	checkAddError(os.Chdir(fluxRepo))

	sourceName := filepath.Join(fluxRepo, fluxRepoName+"-source-"+repoName+".yaml")
	kustName := filepath.Join(fluxRepo, fluxRepoName+"-kustomize-"+repoName+".yaml")
	ioutil.WriteFile(sourceName, source, 0644)
	ioutil.WriteFile(kustName, kust, 0644)
	commitAndPush(sourceName, kustName)

	fmt.Printf("Successfully added repository: %s.\n", repoName)
}
