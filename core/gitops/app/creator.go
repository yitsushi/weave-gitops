package app

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
	"github.com/weaveworks/weave-gitops/core/repository"
	"k8s.io/apimachinery/pkg/util/uuid"
)

type Creator interface {
	Create(repo *git.Repository, name, namespace, description string) (types.App, error)
}

func NewCreator(committerSvc repository.Committer) Creator {
	return &appCreator{
		committerSvc: committerSvc,
	}
}

type appCreator struct {
	committerSvc repository.Committer
}

func (a appCreator) Create(repo *git.Repository, name, namespace, description string) (types.App, error) {
	app := types.App{
		Id:          string(uuid.NewUUID()),
		Name:        name,
		Namespace:   namespace,
		Description: description,
	}

	files, err := app.Files()
	if err != nil {
		return types.App{}, fmt.Errorf("issue creating app files: %w", err)
	}

	commitMessage := fmt.Sprintf("Created new app: %s", app.Name)

	_, err = a.committerSvc.Commit(repo, commitMessage, files)
	if err != nil {
		return types.App{}, fmt.Errorf("git writer failed for app: %w", err)
	}

	return app, nil
}
