package types

import (
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/weaveworks/weave-gitops/core/repository"
	"sigs.k8s.io/kustomize/api/types"
)

type GitopsToolkit struct {
	ClusterName       string
	path              string
	syncRepo          v1beta1.GitRepository
	syncKustomization v1beta2.Kustomization
	Kustomization     types.Kustomization
}

func NewGitopsToolkit(path string, files []repository.File) (GitopsToolkit, error) {

	// We have already pulled the files from the directory or archived source controller
	//

	return GitopsToolkit{}, nil
}

func (gt GitopsToolkit) Namespace() string {
	return gt.syncRepo.ObjectMeta.Namespace
}

func (gt GitopsToolkit) Files() ([]repository.File, error) {
	var files []repository.File

	//kustomizeData, err := yaml.Marshal(gt.Kustomization)
	//if err != nil {
	//	return nil, fmt.Errorf("could not marsha gitops tookit %s into yaml: %w", gt.ClusterName, err)
	//}

	//files = append(files, repository.File{Path: gt.path(kustomizationFilename), Data: kustomizeData})

	return files, nil
}
