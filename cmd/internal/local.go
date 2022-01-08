package internal

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
)

func LocalBranch() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to determine local directory: %w", err)
	}

	repo, err := git.PlainOpen(dir)
	if err == git.ErrRepositoryNotExists {
		return "", err
	} else if err != nil {
		return "", fmt.Errorf("unable to get git repo: %w", err)
	}

	headRef, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("unable to get git ref: %w", err)
	}

	refName := headRef.Name()
	if refName.IsBranch() {
		return refName.String(), nil
	}

	return "", errors.New("current git ref is not a branch")
}
