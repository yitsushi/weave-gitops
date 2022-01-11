package repository

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

const (
	limit = 5
)

var (
	localCache = make(map[string]localRepo)

	ErrBranchDoesNotExist = errors.New("branch does not exist")
)

type localRepo struct {
	dir  string
	repo *git.Repository
}

type Manager interface {
	Get(ctx context.Context, auth transport.AuthMethod, sourceUrl string, branch string) (*git.Repository, error)
	GetTempDir(branch string) (string, error)
}

type repoManager struct {
	cacheMutex sync.Mutex
}

func NewRepoManager() Manager {
	return &repoManager{
		cacheMutex: sync.Mutex{},
	}
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

	// TODO: the temp dir is never removed
	rm.cacheMutex.Lock()
	defer rm.cacheMutex.Unlock()

	localCache[branch] = localRepo{
		dir:  repoDir,
		repo: repo,
	}

	return repo, nil
}

func (rm *repoManager) GetTempDir(branch string) (string, error) {
	rm.cacheMutex.Lock()
	defer rm.cacheMutex.Unlock()

	if lRepo, ok := localCache[branch]; ok {
		return lRepo.dir, nil
	} else {
		return "", ErrBranchDoesNotExist
	}
}
