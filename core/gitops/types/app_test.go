package types

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

func TestAppNameFromPath_NotAnAppPath(t *testing.T) {
	f := setUpAppTest(t)

	appName := appNameFromPath(baseDir)
	f.Expect(appName).To(Equal(""))

	appName = appNameFromPath(baseDir + "/apps_r_us/my-app/test.yaml")
	f.Expect(appName).To(Equal(""), "app name should be an empty string with an invalid path")
}

func TestAppNameFromPath_AnAppPathNoAppSubdirectory(t *testing.T) {
	f := setUpAppTest(t)
	appName := appNameFromPath(baseDir + "/apps")

	f.Expect(appName).To(Equal(""))
}

func TestAppNameFromPath_ValidPaths(t *testing.T) {
	f := setUpAppTest(t)
	appName := appNameFromPath(baseDir + "/apps/my-app-1")
	f.Expect(appName).To(Equal("my-app-1"))

	appName = appNameFromPath(baseDir + "/apps/my-app-2/test.yaml")
	f.Expect(appName).To(Equal("my-app-2"))
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
	f.Expect(files[0].Path).To(Equal(app.path(appFilename)))
	f.Expect(files[1].Path).To(Equal(app.path(kustomizationFilename)))
}
