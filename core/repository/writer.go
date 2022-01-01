package repository

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	repository "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/weaveworks/weave-gitops/pkg/git"
)

const (
	clientName  = "Weave Gitops"
	clientEmail = "weave-gitops@weave.works"
)

type File struct {
	Path string
	Data []byte
}

type GitWriter interface {
	AddCommitAndPush(ctx context.Context, branch, commitMessage string, files []File) error
}

func NewGitWriter(gitClient git.Git, repo repository.GitRepository) GitWriter {
	return &defaultGitWriter{
		gitClient: gitClient,
		repo:      repo,
	}
}

type defaultGitWriter struct {
	gitClient git.Git
	repo      repository.GitRepository
}

func (d defaultGitWriter) AddCommitAndPush(ctx context.Context, branch, commitMessage string, files []File) error {
	repoDir, err := ioutil.TempDir("", "repo-")
	if err != nil {
		return fmt.Errorf("failed creating temp. directory to clone repo: %w", err)
	}

	_, err = d.gitClient.Clone(ctx, repoDir, d.repo.Spec.URL, branch)
	if err != nil {
		return fmt.Errorf("failed cloning repo: %s: %w", d.repo.Spec.URL, err)
	}

	defer os.RemoveAll(repoDir)

	for _, file := range files {
		if err := d.gitClient.Write(file.Path, file.Data); err != nil {
			return fmt.Errorf("failed to write files: %w", err)
		}
	}

	_, err = d.gitClient.Commit(git.Commit{
		Author:  git.Author{Name: clientName, Email: clientEmail},
		Message: commitMessage,
	})

	if err != nil && err != git.ErrNoStagedFiles {
		return fmt.Errorf("failed to update the repository: %w", err)
	}

	if err = d.gitClient.Push(ctx); err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}

	return nil
}
