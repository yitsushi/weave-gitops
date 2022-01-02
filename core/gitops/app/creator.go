package app

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/core/gitops/types"
	"github.com/weaveworks/weave-gitops/core/repository"
	"k8s.io/apimachinery/pkg/util/uuid"
)

type Creator interface {
	Create(name, namespace, description string) (types.App, error)
}

func NewCreator(gitService repository.GitWriter) Creator {
	return &appCreator{
		gitService: gitService,
	}
}

type appCreator struct {
	gitService repository.GitWriter
}

func (d appCreator) Create(name, namespace, description string) (types.App, error) {
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

	commitMessage := fmt.Sprintf("Created new app %s", app.Name)
	err = d.gitService.AddCommitAndPush(context.Background(), "delta", commitMessage, files)
	if err != nil {
		return types.App{}, fmt.Errorf("git writer failed for app: %w", err)
	}

	return app, nil
}
