package types

import (
	"path/filepath"
	"testing"

	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	. "github.com/onsi/gomega"
)

type sourceFixture struct {
	*GomegaWithT
}

func setUpSourceTest(t *testing.T) appFixture {
	return appFixture{
		GomegaWithT: NewGomegaWithT(t),
	}
}

func TestObjectPath(t *testing.T) {
	f := setUpSourceTest(t)
	path := componentFilePath("test/this", ObjectKey{Name: "fake", Namespace: "object"}, v1beta2.KustomizationKind)

	f.Expect(path).To(Equal(filepath.Join("test/this", "fake-object-kustomization.yaml")))
}
