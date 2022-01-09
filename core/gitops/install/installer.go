package install

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
	"github.com/weaveworks/weave-gitops/core/repository"
	"github.com/weaveworks/weave-gitops/manifests"
)

type Installer interface {
	Install(repo *git.Repository) error
}

type gitopsInstall struct {
	committerSvc repository.Committer
}

func NewGitopsInstaller(adder repository.Committer) Installer {
	return &gitopsInstall{
		committerSvc: adder,
	}
}

func (gi gitopsInstall) Install(repo *git.Repository) error {
	files, err := manifests.GitopsManifests(types.BaseDir, manifests.Params{
		AppVersion: "test",
		Namespace:  types.FluxNamespace,
	})
	if err != nil {
		return fmt.Errorf("unable to produce manifest files for Weave Gitops: %w", err)
	}

	_, err = gi.committerSvc.Commit(repo, "Installed Weave GitOps", files)
	if err != nil {
		return fmt.Errorf("there was an issue creating a commit: %w", err)
	}

	return nil
}
