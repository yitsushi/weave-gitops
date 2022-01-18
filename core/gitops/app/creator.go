package app

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
	"github.com/weaveworks/weave-gitops/core/repository"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/rest"
)

type CreateInput struct {
	Name        string
	Namespace   string
	Description string
	DisplayName string
}

type Creator interface {
	Create(repo *git.Repository, auth transport.AuthMethod, input CreateInput) (types.App, error)
}

func NewCreator(committerSvc repository.Writer) Creator {
	return &appCreator{
		committerSvc: committerSvc,
	}
}

type appCreator struct {
	committerSvc repository.Writer
}

func (a appCreator) Create(repo *git.Repository, auth transport.AuthMethod, input CreateInput) (types.App, error) {
	app := types.App{
		Id:          string(uuid.NewUUID()),
		Name:        input.Name,
		Namespace:   input.Namespace,
		Description: input.Description,
		DisplayName: input.DisplayName,
	}

	files, err := app.Files()
	if err != nil {
		return types.App{}, fmt.Errorf("issue creating app files: %w", err)
	}

	commitMessage := fmt.Sprintf("Created new app: %s", app.Name)

	_, err = a.committerSvc.Commit(repo, auth, commitMessage, files)
	if err != nil {
		return types.App{}, fmt.Errorf("git writer failed for app: %w", err)
	}

	return app, nil
}

type KubeCreator interface {
	Create(ctx context.Context, client *rest.RESTClient, input CreateInput) (v1alpha1.Application, error)
}

func NewKubeCreator() Creator {

}

type appKubeCreator struct {
	committerSvc repository.Writer
}

func (a appKubeCreator) Create(ctx context.Context, client *rest.RESTClient, input CreateInput) (types.App, error) {
	app := types.App{
		Id:          string(uuid.NewUUID()),
		Name:        input.Name,
		Namespace:   input.Namespace,
		Description: input.Description,
		DisplayName: input.DisplayName,
	}

	cr := app.CustomResource()

	result := &v1alpha1.Application{}
	client.Post().
		Namespace(cr.Namespace).
		Resource("application").
		Name(input.Name).
		Body(cr).
		Do(ctx).
		Into(result)

	return app, nil
}
