package source

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	repository "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
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

var (
	ErrNotFound = errors.New("not found")
)

type FileJson struct {
	Path string                 `json:"path"`
	Data map[string]interface{} `json:"data"`
}

type Service interface {
	Get(ctx context.Context, name, namespace string) (repository.GitRepository, error)
	GetArtifact(ctx context.Context, name, namespace string) ([]FileJson, error)
	GitClient(ctx context.Context, namespace string, repository repository.GitRepository) (git.Git, error)
}

type defaultService struct {
	client        k8s.Client
	exclusionList []string
}

func NewService(client k8s.Client, exclusionList []string) Service {
	return &defaultService{client: client, exclusionList: exclusionList}
}

func (gr *defaultService) Get(ctx context.Context, name, namespace string) (repository.GitRepository, error) {
	var repoObj repository.GitRepository
	err := gr.client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, &repoObj)

	if apierrors.IsNotFound(err) {
		return repository.GitRepository{}, ErrNotFound
	} else if err != nil {
		return repository.GitRepository{}, err
	}

	return repoObj, nil
}

func (gr *defaultService) GetArtifact(ctx context.Context, name, namespace string) ([]FileJson, error) {
	repo, err := gr.Get(ctx, name, namespace)
	if err != nil {
		return nil, err
	}

	// download the tarball
	//parsedUrl, _ := url.Parse(repo.GetArtifact().URL)
	//parsedUrl.Host = "localhost:8082"
	fmt.Println("This is James")
	req, err := http.NewRequest(http.MethodGet, repo.GetArtifact().URL, nil)
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

	var delta []FileJson

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
			//fmt.Printf("Contents of %s: ", header.Name)
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

			delta = append(delta, FileJson{
				Path: header.Name,
				Data: obj,
			})

		}
	}
}

func (gr *defaultService) GitClient(ctx context.Context, namespace string, repository repository.GitRepository) (git.Git, error) {
	secret := &corev1.Secret{}
	if err := gr.client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      repository.Spec.SecretRef.Name,
	}, secret); apierrors.IsNotFound(err) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("error getting deploy key secret: %w", err)
	}

	pemBytes := auth.ExtractPrivateKey(secret)

	pubKey, err := ssh.NewPublicKeys("git", pemBytes, "")
	if err != nil {
		return nil, fmt.Errorf("could not create public key from secret: %w", err)
	}

	// Set the git client to use the existing deploy key.
	return git.New(pubKey, wrapper.NewGoGit()), nil
}
