package app

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/core/gitops/types"
	"github.com/weaveworks/weave-gitops/core/repository"
)

type Remover interface {
	Remove(app types.App, branch string) error
}

func NewRemover(gitWriter repository.GitWriter) Remover {
	return &appRemover{
		gitWriter: gitWriter,
	}
}

type appRemover struct {
	gitWriter repository.GitWriter
}

func (a appRemover) Remove(app types.App, branch string) error {
	files, err := app.Files()
	if err != nil {
		return fmt.Errorf("issue creating app files: %w", err)
	}

	commitMessage := fmt.Sprintf("Removed app: %s", app.Name)

	err = a.gitWriter.RemoveCommitAndPush(context.Background(), branch, commitMessage, files)
	if err != nil {
		return fmt.Errorf("git writer failed for app: %w", err)
	}

	return nil
}
