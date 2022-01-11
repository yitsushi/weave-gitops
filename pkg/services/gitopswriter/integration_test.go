package gitopswriter

import (
	"context"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"github.com/weaveworks/weave-gitops/pkg/services/gitrepo"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

var _ = Describe("GitopsWriter integration test", func() {
	Describe("RemoveApplication", func() {
		It("removes an Application from a repo", func() {
			auth := &http.BasicAuth{Username: "test-user", Password: "test-password"}
			gitServer, err := testutils.SetupGitServer(auth)
			Expect(err).NotTo(HaveOccurred())
			defer gitServer.StopHTTP()

			repoPath := "bar/test-reponame"
			branchName := "master"

			Expect(gitServer.InitRepo("testdata/git/repo1", branchName, repoPath)).To(Succeed())

			addr := gitServer.HTTPAddressWithCredentials() + "/" + repoPath

			Expect(err).NotTo(HaveOccurred())

			gg := wrapper.NewGoGit()

			gc := git.New(auth, gg)

			l := &loggerfakes.FakeLogger{}
			gp := &gitprovidersfakes.FakeGitProvider{}
			gp.GetDefaultBranchReturns(branchName, nil)

			f := flux.New(osys.New(), &testutils.LocalFluxRunner{})
			auto := automation.NewAutomationGenerator(gp, f, l)

			repoURL, err := gitproviders.NewTestRepoURL(addr, gitproviders.GitProviderGitHub)

			gc2 := gitrepo.NewRepoWriter(repoURL, gp, gc, l)
			writer := NewGitOpsDirectoryWriter(auto, gc2, osys.New(), l)

			Expect(err).NotTo(HaveOccurred())

			app := models.Application{
				Name:       "my-app",
				ConfigRepo: repoURL,
			}

			Expect(writer.RemoveApplication(context.Background(), app, "my-cluster", true)).To(Succeed())

			dir, err := os.MkdirTemp("", "clone-")
			Expect(err).NotTo(HaveOccurred())
			repo, err := gogit.PlainClone(dir, false, &gogit.CloneOptions{URL: addr})
			Expect(err).NotTo(HaveOccurred())

			objs, err := repo.CommitObjects()
			Expect(err).NotTo(HaveOccurred())
			c, err := objs.Next()
			Expect(err).NotTo(HaveOccurred())

			Expect(c.Author.Name).To(Equal("Weave Gitops"))

			files, err := getFiles(dir)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(files)).To(Equal(1))

			k := &types.Kustomization{}

			b, ok := files["/clusters/my-cluster/user/kustomization.yaml"]
			Expect(ok).To(BeTrue())

			Expect(yaml.Unmarshal(b, k)).To(Succeed())

			Expect(k.Resources).To(HaveLen(0))
			Expect(files["/apps/my-app/app.yaml"]).To(BeNil())

		})
	})

	Describe("RemoveClusterRecord", func() {

	})

})

func getFiles(dir string) (map[string][]byte, error) {
	files := map[string][]byte{}

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() && strings.Contains(path, ".weave-gitops") {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			a := strings.Split(path, ".weave-gitops")
			files[a[1]] = b
		}

		return nil
	})

	return files, err
}
