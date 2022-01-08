package install

import (
	"testing"

	. "github.com/onsi/gomega"
)

const (
	testNamespace = "test-system"
)

type installerFixture struct {
	*GomegaWithT
}

func setUpInstallerTest(t *testing.T) installerFixture {
	return installerFixture{
		GomegaWithT: NewGomegaWithT(t),
	}
}
