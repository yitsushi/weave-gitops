package add

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/git"
)

//go:embed manifests/workload/image.yaml
var embeddedImageManifest []byte

func runCmd2(cmd *cobra.Command, args []string) {
	if params.url == "" {
		checkAddError(fmt.Errorf("url is required"))
	}

	parts := strings.Split(params.url, "/")

	owner := parts[len(parts)-2]
	workloadRepoName := strings.ReplaceAll(parts[len(parts)-1], ".git", "")

	err := git.PushChangesChangesToWegoRepo(owner, params.infraRepo, func(repoDir string) string {
		clusterDir := path.Join(repoDir, "clusters", "my-cluster")

		// Writing flux manifests: kustomization and source
		err := writeFluxManifests(clusterDir, workloadRepoName)
		checkError("failed to generate kpack manifests", err)

		// Writing kpack manifests: builder and image
		err = writeKpackEmbeddedManifests(clusterDir, owner, workloadRepoName)
		checkError("failed to generate kpack manifests", err)

		return fmt.Sprintf("Add %s/%s workload manifests", owner, workloadRepoName)
	})
	checkError("failed to push changes to github", err)
}

func writeFluxManifests(clusterDir string, repoName string) error {
	workloadDir := path.Join(clusterDir, "workloads", repoName)
	err := os.MkdirAll(workloadDir, os.ModePerm)
	if err != nil {
		return err
	}

	sourceManifest := generateSourceManifest(repoName)
	err = os.WriteFile(path.Join(workloadDir, "source.yaml"), sourceManifest, 0666)
	if err != nil {
		return err
	}

	kustomizeManifest := generateKustomizeManifest(repoName)
	err = os.WriteFile(path.Join(workloadDir, "kustomize.yaml"), kustomizeManifest, 0666)
	if err != nil {
		return err
	}

	return nil
}

func writeKpackEmbeddedManifests(clusterDir string, owner string, repoName string) error {
	workloadDir := path.Join(clusterDir, "workloads", repoName)
	err := os.MkdirAll(workloadDir, os.ModePerm)
	if err != nil {
		return err
	}

	imageManifest := generateImageManifest(Workload{
		Name:      repoName,
		DockerTag: fmt.Sprintf("%s/%s", owner, repoName),
		GitURL:    params.url,
	})
	err = os.WriteFile(path.Join(workloadDir, "image.yaml"), imageManifest, 0666)
	if err != nil {
		return err
	}

	return nil
}

// func writeKpackEmbeddedManifests(clusterDir string) error {
// 	embeddedDir := "manifests/bootstrap"
// 	manifests, err := fs.ReadDir(embeddedManifests, embeddedDir)
// 	if err != nil {
// 		return err
// 	}
// 	for _, manifest := range manifests {
// 		if manifest.IsDir() {
// 			continue
// 		}

// 		data, err := fs.ReadFile(embeddedManifests, path.Join(embeddedDir, manifest.Name()))
// 		if err != nil {
// 			return fmt.Errorf("reading file failed: %w", err)
// 		}

// 		kpackDir := path.Join(clusterDir, "kpack")
// 		os.MkdirAll(kpackDir, os.ModePerm)

// 		err = os.WriteFile(path.Join(kpackDir, manifest.Name()), data, 0666)
// 		if err != nil {
// 			return fmt.Errorf("writing file failed: %w", err)
// 		}
// 	}
// 	return nil
// }

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
