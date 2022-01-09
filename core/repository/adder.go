package repository

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Committer interface {
	Commit(repo *git.Repository, msg string, files []File) (string, error)
}

type gitCommitter struct {
}

func NewGitCommitter() Committer {
	return &gitCommitter{}
}

func (g gitCommitter) Commit(repo *git.Repository, msg string, files []File) (string, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("unable to get worktree from repo: %w", err)
	}

	for _, file := range files {
		f, err := worktree.Filesystem.Create(file.Path)
		if err != nil {
			return "", fmt.Errorf("failed to create file in %s: %w", file.Path, err)
		}

		_, err = io.Copy(f, bytes.NewReader(file.Data))

		f.Close()

		err = worktree.AddWithOptions(&git.AddOptions{Path: file.Path})
		if err != nil {
			return "", fmt.Errorf("unable to stage file: %s", err)
		}
	}

	commit, err := worktree.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  ClientName,
			Email: ClientEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("unable to commit: %w", err)
	}

	return commit.String(), nil
}
