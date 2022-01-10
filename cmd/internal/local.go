package internal

import (
	"fmt"
	"io/ioutil"

	"github.com/go-git/go-git/v5"
	"github.com/weaveworks/weave-gitops/core/repository"
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

func ReadDir(dir string) ([]repository.File, error) {
	fileInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("")
	}

	var files []repository.File
	for _, fi := range fileInfo {
		if fi.IsDir() {
			continue
		}

		data, readErr := ioutil.ReadFile(fi.Name())
		if readErr != nil {
			return nil, fmt.Errorf("error reading file: %s", fi.Name())
		}

		files = append(files, repository.File{
			Path: fi.Name(),
			Data: data,
		})
	}

	return files, nil
}
