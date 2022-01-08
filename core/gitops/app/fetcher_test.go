package app

import (
	"testing"

	. "github.com/onsi/gomega"
)

type fetcherFixture struct {
	*GomegaWithT
}

func setUpFetcherTest(t *testing.T) fetcherFixture {
	return fetcherFixture{
		GomegaWithT: NewGomegaWithT(t),
	}
}
