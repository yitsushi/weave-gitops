package app

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
	"github.com/weaveworks/weave-gitops/core/repository"
)

type Remover interface {
	Remove(repo *git.Repository, auth transport.AuthMethod, appName, namespace, repoName string) error
}

func NewRemover(commitSvc repository.Committer, fetcher Fetcher) Remover {
	return &appRemover{
		commitSvc: commitSvc,
		fetcher:   fetcher,
	}
}

type appRemover struct {
	commitSvc repository.Committer
	fetcher   Fetcher
}

func (a appRemover) Remove(repo *git.Repository, auth transport.AuthMethod, appName, namespace, repoName string) error {
	appObj, err := a.fetcher.Get(context.Background(), appName, repoName, namespace)
	if err == types.ErrNotFound {
		return nil
	} else if err != nil {
		return fmt.Errorf("issue getting an application: %w", err)
	}

	files, err := appObj.Files()
	if err != nil {
		return fmt.Errorf("issue creating app files: %w", err)
	}

	commitMessage := fmt.Sprintf("Removed app: %s", appName)

	_, err = a.commitSvc.Commit(repo, auth, commitMessage, files)
	if err != nil {
		return fmt.Errorf("git writer failed for app: %w", err)
	}

	return nil
}
