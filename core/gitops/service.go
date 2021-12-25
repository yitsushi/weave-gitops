package gitops

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/util/uuid"
)

type AppService interface {
	Create(name, namespace, description string) (App, error)
	Get(name string) App
}

func NewAppService(gitService GitService) AppService {
	return &defaultAppService{
		gitService: gitService,
	}
}

type defaultAppService struct {
	gitService GitService
}

func (d defaultAppService) Create(name, namespace, description string) (App, error) {
	app := App{
		Id:          string(uuid.NewUUID()),
		Name:        name,
		Namespace:   namespace,
		Description: description,
	}

	files, err := app.Files()
	if err != nil {
		return App{}, fmt.Errorf("issue creating app files: %w", err)
	}

	commitMessage := fmt.Sprintf("Created new app %s", app.Name)
	d.gitService.AddCommitAndPush(context.Background(), "delta", commitMessage, files)

	return app, nil
}

func (d defaultAppService) Get(name string) App {
	panic("implement me")
}
