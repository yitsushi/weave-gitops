package kustomize

import (
	"context"
	"fmt"

	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/weaveworks/weave-gitops/core/gitops/app"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
	"github.com/weaveworks/weave-gitops/core/repository"
	"k8s.io/client-go/rest"
)

const (
	kustomizations = "kustomizations"
)

type CreateInput struct {
	AppName       string
	Kustomization v1beta2.Kustomization
}

type Creator interface {
	Create(dir string, repo *git.Repository, auth transport.AuthMethod, input CreateInput) (v1beta2.Kustomization, error)
}

func NewCreator(writer repository.Writer, fetcher app.RepoFetcher) Creator {
	return &kustomizeCreator{
		writer:      writer,
		repoFetcher: fetcher,
	}
}

type kustomizeCreator struct {
	writer      repository.Writer
	repoFetcher app.RepoFetcher
}

func (a kustomizeCreator) Create(dir string, repo *git.Repository, auth transport.AuthMethod, input CreateInput) (v1beta2.Kustomization, error) {
	app, err := a.repoFetcher.Get(dir, input.AppName)
	if err == types.ErrNotFound {
		return v1beta2.Kustomization{}, err
	} else if err != nil {
		return v1beta2.Kustomization{}, fmt.Errorf("kustServer.Add: %w", err)
	}

	files, err := app.AddFluxKustomization(input.Kustomization)
	if err != nil {
		return v1beta2.Kustomization{}, fmt.Errorf("kustomizeCreate: issue creating app files: %w", err)
	}

	commitMessage := fmt.Sprintf("New kustomization %s in app %s", input.Kustomization.ObjectMeta.Name, app.Name)

	_, err = a.writer.Commit(repo, auth, commitMessage, files)
	if err != nil {
		return v1beta2.Kustomization{}, fmt.Errorf("git writer failed for app: %w", err)
	}

	return input.Kustomization, nil
}

type KubeCreator interface {
	Create(ctx context.Context, client *rest.RESTClient, kustomization *v1beta2.Kustomization) (result *v1beta2.Kustomization, err error)
}

func NewK8sCreator() KubeCreator {
	return &kustomizationCreator{}
}

type kustomizationCreator struct {
}

func (a kustomizationCreator) Create(ctx context.Context, client *rest.RESTClient, kustomization *v1beta2.Kustomization) (result *v1beta2.Kustomization, err error) {
	result = &v1beta2.Kustomization{}
	err = client.Post().
		Namespace(kustomization.ObjectMeta.Namespace).
		Resource(kustomizations).
		Name(kustomization.ObjectMeta.Name).
		Body(kustomization).
		Do(ctx).
		Into(result)

	return
}
