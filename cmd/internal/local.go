package internal

import (
	"fmt"

	"github.com/go-git/go-git/v5"
)

func GitRepository(dir string) (*git.Repository, error) {
	repo, err := git.PlainOpen(dir)
	if err == git.ErrRepositoryNotExists {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("unable to get git repo: %w", err)
	}

	return repo, nil
}
