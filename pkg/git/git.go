package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/k0kubun/pp"
)

func PushChangesChangesToWegoRepo(owner string, repo string, changes func(repoDir string) string) error {
	tmpDir, err := ioutil.TempDir("", "wego-repo-")
	// defer os.RemoveAll(tmpDir)
	if err != nil {
		return err
	}
	pp.Println(tmpDir)

	publicKeys, err := ssh.NewPublicKeysFromFile("git", filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"), "")
	if err != nil {
		return err
	}

	repository := fmt.Sprintf("git@github.com:%s/%s.git", owner, repo)
	r, err := git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:  repository,
		Auth: publicKeys,
	})
	if err != nil {
		return err
	}
	fmt.Printf("%s repository cloned\n", repository)

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	// executing repo changes
	commitMsg := changes(tmpDir)

	// pushing manifests to the repo
	err = w.AddGlob(".")
	if err != nil {
		return err
	}

	_, err = w.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Wego CLI",
			Email: "wego@weave.works",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	err = r.Push(&git.PushOptions{
		Auth: publicKeys,
	})
	if err != nil {
		return err
	}

	return nil
}
