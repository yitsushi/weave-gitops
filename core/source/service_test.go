package source

import (
	"context"
	"testing"

	repository "github.com/fluxcd/source-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testNamespace = "test-system"
)

type repoFixture struct {
	*GomegaWithT
	k8s client.WithWatch
}

func setUpRepoTest(t *testing.T) repoFixture {
	return repoFixture{
		GomegaWithT: NewGomegaWithT(t),
		k8s:         fake.NewClientBuilder().WithScheme(kube.CreateScheme()).Build(),
	}
}

func TestRepoDoesNotExist(t *testing.T) {
	f := setUpRepoTest(t)
	client := NewService(f.k8s, []string{})

	repo, err := client.Get(context.Background(), "gitops-repo", testNamespace)

	f.Expect(err).To(MatchError(ErrNotFound))
	f.Expect(repo).To(Equal(repository.GitRepository{}))
}

func TestRepoExists(t *testing.T) {
	f := setUpRepoTest(t)
	client := NewService(f.k8s, []string{})

	repoObj := repository.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-repo",
			Namespace: "test-system",
		},
	}
	f.k8s.Create(context.Background(), &repoObj)

	repo, err := client.Get(context.Background(), "gitops-repo", testNamespace)

	f.Expect(err).To(BeNil())
	f.Expect(repo.Name).To(Equal("gitops-repo"))
	f.Expect(repo.Namespace).To(Equal(testNamespace))
}
