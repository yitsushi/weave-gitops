package types

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/weaveworks/weave-gitops/core/repository"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/kustomize/api/types"
	k8yaml "sigs.k8s.io/yaml"
)

const (
	gotkSyncFileName  = "gotk-sync.yaml"
	kindGitRepository = "GitRepository"
	kindKustomization = "Kustomization"
	systemPath        = "system"
)

var (
	ErrNoGitopsToolkitFiles = errors.New("gitops toolkit files not found")
)

type GitopsToolkit struct {
	ClusterName       string
	SystemPath        string
	syncRepo          v1beta1.GitRepository
	syncKustomization v1beta2.Kustomization
	Kustomization     types.Kustomization
}

func NewGitopsToolkit(files []repository.File) (GitopsToolkit, error) {

	if len(files) == 0 {
		return GitopsToolkit{}, ErrNoGitopsToolkitFiles
	}

	toolkit := GitopsToolkit{}

	for _, file := range files {
		if !strings.HasSuffix(file.Path, gotkSyncFileName) {
			continue
		}

		r := yaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(file.Data)))

		for {
			data, err := r.Read()
			if err == io.EOF {
				break
			}

			var gitRepository v1beta1.GitRepository

			err = yaml.Unmarshal(data, &gitRepository)
			if err != nil {
				return GitopsToolkit{}, fmt.Errorf("gitops toolkit cannot unmarshal into gitRepository from %s: %w", gotkSyncFileName, err)
			}

			if gitRepository.Kind == kindGitRepository {
				toolkit.syncRepo = gitRepository
				continue
			}

			var kustomization v1beta2.Kustomization

			err = yaml.Unmarshal(data, &kustomization)
			if err != nil {
				return GitopsToolkit{}, fmt.Errorf("gitops toolkit cannot unmarshal into kustomization from %s: %w", gotkSyncFileName, err)
			}

			if kustomization.Kind == kindKustomization {
				toolkit.syncKustomization = kustomization
				continue
			}
		}
	}

	toolkit.SystemPath = filepath.Join(toolkit.syncKustomization.Spec.Path, systemPath)
	toolkit.Kustomization = types.Kustomization{
		TypeMeta: types.TypeMeta{
			Kind:       kindKustomization,
			APIVersion: types.KustomizationVersion,
		},
		Resources: []string{
			"./gitops/app",
			"./gitops/runtime",
		},
	}

	return toolkit, nil
}

func (gt GitopsToolkit) Namespace() string {
	return gt.syncRepo.ObjectMeta.Namespace
}

func (gt GitopsToolkit) Files() ([]repository.File, error) {
	var files []repository.File

	kustomizeData, err := k8yaml.Marshal(gt.Kustomization)
	if err != nil {
		return nil, fmt.Errorf("could not marsha gitops tookit %s into yaml: %w", gt.ClusterName, err)
	}

	files = append(files, repository.File{Path: filepath.Join(gt.SystemPath, kustomizationFilename), Data: kustomizeData})

	return files, nil
}
