package app

import (
	"testing"

	. "github.com/onsi/gomega"
)

type removerFixture struct {
	*GomegaWithT
}

func setUpRemoverTest(t *testing.T) removerFixture {
	return removerFixture{
		GomegaWithT: NewGomegaWithT(t),
	}
}
