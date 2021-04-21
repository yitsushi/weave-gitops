package add

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/k0kubun/pp"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

//go:embed manifests/bootstrap/*.yaml
var embeddedManifests embed.FS

//go:embed manifests/workload/image.yaml
var embeddedImageManifest []byte

func runCmd2(cmd *cobra.Command, args []string) {
	if params.url == "" {
		checkAddError(fmt.Errorf("url is required"))
	}

	parts := strings.Split(params.url, "/")

	owner := parts[len(parts)-2]
	repoName := strings.ReplaceAll(parts[len(parts)-1], ".git", "")

	// fluxArgs := []string{
	// 	"--branch=main",
	// 	"--personal=true",
	// 	"--components-extra=image-reflector-controller,image-automation-controller",
	// 	"--read-write-key=true",
	// 	fmt.Sprintf("--owner=%s", owner),
	// 	fmt.Sprintf("--repository=%s", "wego-infra"),
	// 	fmt.Sprintf("--path=%s", "./clusters/my-cluster"),
	// }
	// _, err := fluxops.CallFlux(fmt.Sprintf("bootstrap github %s", strings.Join(fluxArgs, " ")))
	// checkAddError(err)

	pushChangesChangesToWegoRepo(func(repoDir string) string {
		clusterDir := path.Join(repoDir, "clusters", "my-cluster")

		writeFluxWorkflowManifests(clusterDir, repoName)
		writeKpackEmbeddedManifests(clusterDir)
		writeWorkloadEmbeddedManifests(clusterDir, owner, repoName)

		return "pushing image manifest"
	})
}

func writeFluxWorkflowManifests(clusterDir string, repoName string) {
	workloadDir := path.Join(clusterDir, "workloads", repoName)
	err := os.MkdirAll(workloadDir, os.ModePerm)
	checkAddError(err)

	sourceManifest := generateSourceManifest(repoName)
	err = os.WriteFile(path.Join(workloadDir, "source.yaml"), sourceManifest, 0666)
	checkAddError(err)

	kustomizeManifest := generateKustomizeManifest(repoName)
	err = os.WriteFile(path.Join(workloadDir, "kustomize.yaml"), kustomizeManifest, 0666)
	checkAddError(err)
}

func pushChangesChangesToWegoRepo(changes func(repoDir string) string) {
	tmpDir, err := ioutil.TempDir("", "wego-repo-")
	defer os.RemoveAll(tmpDir)
	checkAddError(err)
	pp.Println(tmpDir)

	publicKeys, err := ssh.NewPublicKeysFromFile("git", filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"), "")
	checkAddError(err)

	repository := "git@github.com:luizbafilho/wego-infra.git"
	r, err := git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:  repository,
		Auth: publicKeys,
	})
	checkAddError(err)
	fmt.Println("Repository cloned")

	w, err := r.Worktree()
	checkAddError(err)

	// executing repo changes
	commitMsg := changes(tmpDir)

	// pushing manifests to the repo
	err = w.AddGlob(".")
	checkAddError(errors.Wrap(err, "failed to add"))

	_, err = w.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Wego CLI",
			Email: "wego@weave.works",
			When:  time.Now(),
		},
	})
	checkAddError(errors.Wrap(err, "failed to commit"))

	err = r.Push(&git.PushOptions{
		Auth: publicKeys,
	})
	checkAddError(err)
}

func writeWorkloadEmbeddedManifests(clusterDir string, owner string, repoName string) {
	workloadDir := path.Join(clusterDir, "workloads", "podinfo")
	err := os.MkdirAll(workloadDir, os.ModePerm)
	checkAddError(err)

	imageManifest := generateImageManifest(Workload{
		Name:      repoName,
		DockerTag: fmt.Sprintf("%s/%s", owner, repoName),
		GitURL:    params.url,
	})
	err = os.WriteFile(path.Join(workloadDir, "image.yaml"), imageManifest, 0666)
	checkAddError(err)
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

		kpackDir := path.Join(clusterDir, "kpack")
		os.MkdirAll(kpackDir, os.ModePerm)

		err = os.WriteFile(path.Join(kpackDir, manifest.Name()), data, 0666)
		if err != nil {
			return fmt.Errorf("writing file failed: %w", err)
		}
	}
	return nil
}

type Workload struct {
	Name      string
	DockerTag string
	GitURL    string
}

func generateImageManifest(workload Workload) []byte {
	tpl, err := template.New("image").Parse(string(embeddedImageManifest))
	checkAddError(err)

	var b bytes.Buffer
	writter := io.Writer(&b)

	err = tpl.Execute(writter, workload)
	if err != nil {
		panic(err)
	}

	return b.Bytes()
}
