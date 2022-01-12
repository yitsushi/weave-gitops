package repository

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

const (
	limit = 5
)

var (
	localCache map[string]localRepo
)

type localRepo struct {
	pattern string
	repo    *git.Repository
}

type Manager interface {
	Get(ctx context.Context, auth transport.AuthMethod, sourceUrl string, branch string) (*git.Repository, error)
}

type repoManager struct {
}

func NewRepoManager() Manager {
	return &repoManager{}
}

func (rm *repoManager) Get(ctx context.Context, auth transport.AuthMethod, sourceUrl string, branch string) (*git.Repository, error) {
	repoDir, err := ioutil.TempDir("", fmt.Sprintf("repo-%s", branch))
	if err != nil {
		return nil, fmt.Errorf("failed creating temp. directory to clone repo: %w", err)
	}

	branchRef := plumbing.NewBranchReferenceName(branch)
	repo, err := git.PlainCloneContext(ctx, repoDir, false, &git.CloneOptions{
		URL:           sourceUrl,
		Auth:          auth,
		RemoteName:    git.DefaultRemoteName,
		ReferenceName: branchRef,
		SingleBranch:  true,
		NoCheckout:    false,
		Progress:      nil,
		Depth:         0,
		Tags:          git.NoTags,
	})
	if err != nil {
		return nil, fmt.Errorf("failed cloning repo: %s: %w", sourceUrl, err)
	}

	return repo, nil
}
