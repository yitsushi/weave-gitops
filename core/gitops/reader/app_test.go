package reader

import (
	"fmt"
	"path/filepath"
	"testing"
	"testing/fstest"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
)

// TODO: Make sure we can read multiple yamls out of a single file

var (
	appYaml = fmt.Sprintf(`
apiVersion: %s
kind: %s
metadata:
  creationTimestamp: null
  name: app-19
  namespace: wego-system
spec:
  description: This is a description for the app.
status: {}
`, types.ApplicationVersion, types.ApplicationKind)

	kustomizeYaml = `
apiVersion: kustomize.config.k8s.io/v1beta1
commonLabels:
  gitops.weave.works/app-id: 836bac7a-cd97-476b-8989-0d66b5bb82c5
kind: Kustomization
metadata:
  name: app-19
  namespace: wego-system
resources:
- ./app.yaml
`
)

type AppsFixture struct {
	*GomegaWithT
}

func arrangeOpen() fstest.MapFS {
	return fstest.MapFS{
		filepath.Join(types.AppPath("app-1"), types.AppFilename):           {Data: []byte(appYaml)},
		filepath.Join(types.AppPath("app-1"), types.KustomizationFilename): {Data: []byte(kustomizeYaml)},
	}
}

func setUpAppsTest(t *testing.T) AppsFixture {
	return AppsFixture{
		GomegaWithT: NewGomegaWithT(t),
	}
}

func TestAppFromPaths_InvalidPath(t *testing.T) {
	f := setUpAppsTest(t)

	fileSystem := arrangeOpen()
	_, err := ReadApps(fileSystem, []string{".weave-gitops/ka-boom"})

	f.Expect(err).To(MatchError("invalid path for file .weave-gitops/ka-boom"))
}

func TestAppFromPaths_App(t *testing.T) {
	f := setUpAppsTest(t)

	fileSystem := arrangeOpen()
	apps, err := ReadApps(fileSystem, []string{
		".weave-gitops/apps/app-1/app.yaml",
		".weave-gitops/apps/app-1/kustomization.yaml",
	})

	f.Expect(err).To(BeNil())
	f.Expect(apps).To(HaveLen(1))
	f.Expect(apps["app-1"]).To(Equal(types.App{
		Id:          "836bac7a-cd97-476b-8989-0d66b5bb82c5",
		Name:        "app-1",
		Namespace:   "wego-system",
		Description: "This is a description for the app.",
		DisplayName: "",
	}))
}
