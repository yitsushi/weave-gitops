package internal

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

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

func ReadDir(directories []string) ([][]repository.File, error) {
	var toolkitFiles [][]repository.File

	for _, dir := range directories {
		fileInfo, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("could not find dir %s, are you sure you put the correct path: %w", dir, err)
		}

		var files []repository.File

		for _, fi := range fileInfo {
			if fi.IsDir() {
				continue
			}

			path := filepath.Join(dir, fi.Name())

			data, readErr := ioutil.ReadFile(path)
			if readErr != nil {
				return nil, fmt.Errorf("error reading file: %s", fi.Name())
			}

			files = append(files, repository.File{
				Path: path,
				Data: data,
			})
		}

		toolkitFiles = append(toolkitFiles, files)
	}

	return toolkitFiles, nil
}

type StubAuth struct{}

func (StubAuth) String() string {
	return ""
}

func (StubAuth) Name() string {
	return ""
}
