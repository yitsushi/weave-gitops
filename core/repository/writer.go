package repository

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

const (
	ClientName  = "Weave Gitops"
	ClientEmail = "weave-gitops@weave.works"
)

type File struct {
	Path string
	Data []byte
}

type Writer interface {
	Commit(repo *git.Repository, auth transport.AuthMethod, msg string, files []File) (string, error)
}

type gitWriter struct {
	remove bool
	push   bool
}

func NewGitWriter(push bool) Writer {
	return &gitWriter{
		push:   push,
		remove: false,
	}
}

func NewGitDeleter(push bool) Writer {
	return &gitWriter{
		push:   push,
		remove: true,
	}
}

func (g gitWriter) Commit(repo *git.Repository, auth transport.AuthMethod, msg string, files []File) (string, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("unable to get worktree from repo: %w", err)
	}

	for _, file := range files {

		if g.remove {
			err := worktree.Filesystem.Remove(file.Path)
			if err != nil {
				return "", fmt.Errorf("failed to remove file in %s: %w", file.Path, err)
			}
		} else {
			f, err := worktree.Filesystem.Create(file.Path)
			if err != nil {
				return "", fmt.Errorf("failed to create file in %s: %w", file.Path, err)
			}

			_, err = io.Copy(f, bytes.NewReader(file.Data))
			if err != nil {
				return "", fmt.Errorf("failed to copy data to temp file in worktree %s: %w", file.Path, err)
			}

			_ = f.Close()
		}

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

	if g.push {
		if err := repo.Push(&git.PushOptions{
			RemoteName: git.DefaultRemoteName,
			Auth:       auth,
		}); err != nil {
			return "", fmt.Errorf("could not push to remote: %s", err)
		}
	}

	return commit.String(), nil
}
