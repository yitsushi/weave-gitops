package app

import (
	"testing"

	. "github.com/onsi/gomega"
)

const (
	testNamespace = "test-system"
)

type appFixture struct {
	*GomegaWithT
}

func setUpAppTest(t *testing.T) appFixture {
	return appFixture{
		GomegaWithT: NewGomegaWithT(t),
	}
}

func TestAppFilesSuccess(t *testing.T) {
	f := setUpAppTest(t)
	app := App{
		Id:          "12345",
		Name:        "my-app",
		Namespace:   testNamespace,
		Description: "This is my test application, it's going to take over the world",
	}

	files, err := app.Files()
	f.Expect(err).To(BeNil())
	f.Expect(len(files)).To(Equal(2))
	f.Expect(files[0].Path).To(Equal(app.path(kustomizationFilename)))
	f.Expect(files[1].Path).To(Equal(app.path("metadata.json")))
}
