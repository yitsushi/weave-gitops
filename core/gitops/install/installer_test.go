package install

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/repository"
)

const (
	testNamespace = "test-system"
)

type installerFixture struct {
	*GomegaWithT
	committer repository.Writer
}

func setUpInstallerTest(t *testing.T) installerFixture {
	return installerFixture{
		committer:   repository.NewGitCommitter(),
		GomegaWithT: NewGomegaWithT(t),
	}
}

func TestNewGitopsInstaller(t *testing.T) {
	//f := setUpInstallerTest(t)

	//_, err := NewGitopsInstaller(f.committer)
	//
	//f.Expect(err).To(BeNil())
}
