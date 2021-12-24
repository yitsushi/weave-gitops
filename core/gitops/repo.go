package gitops

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
)

type Service interface {
}

func NewService() (Service, error) {
	return &defaultService{}, nil
}

type defaultService struct {
}

func CloneRepo(ctx context.Context, client git.Git, url gitproviders.RepoURL, branch string) (func(), string, error) {
	repoDir, err := ioutil.TempDir("", "user-repo-")
	if err != nil {
		return nil, "", fmt.Errorf("failed creating temp. directory to clone repo: %w", err)
	}

	_, err = client.Clone(ctx, repoDir, url.String(), branch)
	if err != nil {
		return nil, "", fmt.Errorf("failed cloning user repo: %s: %w", url, err)
	}

	return func() {
		_ = os.RemoveAll(repoDir)
	}, repoDir, nil
}
