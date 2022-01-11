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
)

type CreateInput struct {
	AppName       string
	RepoName      string
	Kustomization v1beta2.Kustomization
}

type Creator interface {
	Create(ctx context.Context, repo *git.Repository, auth transport.AuthMethod, input CreateInput) (v1beta2.Kustomization, error)
}

func NewCreator(committerSvc repository.Writer, fetcher app.Fetcher) Creator {
	return &kustomizeCreator{
		committerSvc: committerSvc,
		fetcher:      fetcher,
	}
}

type kustomizeCreator struct {
	committerSvc repository.Writer
	fetcher      app.Fetcher
}

func (a kustomizeCreator) Create(ctx context.Context, repo *git.Repository, auth transport.AuthMethod, input CreateInput) (v1beta2.Kustomization, error) {
	app, err := a.fetcher.Get(ctx, input.AppName, input.RepoName, types.FluxNamespace)
	if err == types.ErrNotFound {
		return v1beta2.Kustomization{}, err
	} else if err != nil {
		return v1beta2.Kustomization{}, fmt.Errorf("kustServer.Add: %w", err)
	}

	files, err := app.Files()
	if err != nil {
		return v1beta2.Kustomization{}, fmt.Errorf("kustomizeCreate: issue creating app files: %w", err)
	}

	commitMessage := fmt.Sprintf("Created new kustomization: %s", app.Name)

	_, err = a.committerSvc.Commit(repo, auth, commitMessage, files)
	if err != nil {
		return v1beta2.Kustomization{}, fmt.Errorf("git writer failed for app: %w", err)
	}

	return input.Kustomization, nil
}
