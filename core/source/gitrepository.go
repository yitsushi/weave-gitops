package source

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	repository "github.com/fluxcd/source-controller/api/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	kFluxSystem = "wego-system"
)

var (
	GitopsRuntimeExclusionList = []string{
		"/system/gitops-runtime.yaml",
		"/system/flux-system-kustomization-resource.yaml",
		"/system/flux-user-kustomization-resource.yaml",
		"/system/wego-app.yaml",
		"/system/wego-config.yaml",
		"/system/wego-system.yaml",
		".keep",
	}
)

type K8sObject struct {
	Path string                 `json:"path"`
	Data map[string]interface{} `json:"data"`
}

type GitRepository interface {
	Get(ctx context.Context) (repository.GitRepository, error)
	GetArtifact(ctx context.Context) ([]K8sObject, error)
}

type gitRepository struct {
	client        client.Client
	exclusionList []string
}

func NewGitRepository(client client.Client, exclusionList []string) GitRepository {
	return &gitRepository{client: client, exclusionList: exclusionList}
}

func (gr *gitRepository) Get(ctx context.Context) (repository.GitRepository, error) {
	var repoObj repository.GitRepository
	err := gr.client.Get(ctx, types.NamespacedName{
		Namespace: kFluxSystem,
		Name:      "wego-github-gitops-repo-000-delta",
	}, &repoObj)
	if err != nil {
		return repository.GitRepository{}, err
	}

	return repoObj, nil
}

func (gr *gitRepository) GetArtifact(ctx context.Context) ([]K8sObject, error) {
	repo, err := gr.Get(ctx)
	if err != nil {
		return nil, err
	}

	// download the tarball
	parsedUrl, _ := url.Parse(repo.GetArtifact().URL)
	parsedUrl.Host = "localhost:8082"
	req, err := http.NewRequest(http.MethodGet, parsedUrl.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request, error: %w", err)
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to download artifact from %s, error: %w", repo.GetArtifact().URL, err)
	}
	defer resp.Body.Close()

	// check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download artifact, status: %s", resp.Status)
	}

	// extract
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	var delta []K8sObject

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return delta, nil

		// return any other error
		case err != nil:
			return nil, err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		include := true
		for _, exclusionPath := range gr.exclusionList {
			if strings.HasSuffix(header.Name, exclusionPath) {
				//fmt.Printf("Excluding contents of %s: ", header.Name)
				include = false
				break
			}
		}

		if include {
			data := bytes.NewBuffer([]byte(nil))
			fmt.Printf("Contents of %s: ", header.Name)
			if _, err := io.Copy(data, tr); err != nil {
				return nil, err
			}

			jsonData, err := yaml.YAMLToJSON(data.Bytes())
			if err != nil {
				fmt.Printf("AN ERROR OCCURRED CONVERTING TO YAML")
				return nil, err
			}

			var obj map[string]interface{}
			err = json.Unmarshal(jsonData, &obj)
			if err != nil {
				fmt.Printf("AN ERROR OCCURRED CONVERTING TO JSON")
				return nil, err
			}

			delta = append(delta, K8sObject{
				Path: header.Name,
				Data: obj,
			})

		}
	}
}
