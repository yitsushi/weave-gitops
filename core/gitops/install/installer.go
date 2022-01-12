package install

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
	"github.com/weaveworks/weave-gitops/core/repository"
	"github.com/weaveworks/weave-gitops/manifests"
)

type Installer interface {
	Install(repo *git.Repository, auth transport.AuthMethod, toolkitFiles [][]repository.File) error
}

type gitopsInstall struct {
	appVersion   string
	committerSvc repository.Committer
}

func NewGitopsInstaller(adder repository.Committer, version string) Installer {
	return &gitopsInstall{
		appVersion:   version,
		committerSvc: adder,
	}
}

func (gi gitopsInstall) Install(repo *git.Repository, auth transport.AuthMethod, toolkitFiles [][]repository.File) error {

	for _, tkf := range toolkitFiles {
		toolkit, err := types.NewGitopsToolkit(tkf)
		if err != nil {
			return fmt.Errorf("unable to create gitops toolkit: %w", err)
		}

		files, err := manifests.GitopsManifests(toolkit.SystemPath, manifests.Params{
			AppVersion: gi.appVersion,
			Namespace:  toolkit.Namespace(),
		})
		if err != nil {
			return fmt.Errorf("unable to produce manifest files for Weave Gitops: %w", err)
		}

		systemFiles, err := toolkit.Files()
		if err != nil {
			return fmt.Errorf("unable to create system files for Weave Gitops: %w", err)
		}

		files = append(files, systemFiles...)

		_, err = gi.committerSvc.Commit(repo, auth, fmt.Sprintf("Installed Weave GitOps in %s", toolkit.ClusterName), files)
		if err != nil {
			return fmt.Errorf("there was an issue creating a commit: %w", err)
		}
	}

	return nil
}
