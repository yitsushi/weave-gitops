package reader

import (
	"fmt"
	"path/filepath"
	"testing"
	"testing/fstest"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
	k8types "sigs.k8s.io/kustomize/api/types"
)

// TODO: Make sure we can read multiple yamls out of a single file

const ()

var (
	testDir = fmt.Sprintf("Users/jack-skeleton/spooky-repo/%s", types.AppPathPrefix)

	appYaml = fmt.Sprintf(`
apiVersion: %s
kind: %s
metadata:
  name: app-1
  namespace: wego-system
spec:
  description: This is a description for the app.
  displayName: App 1
status: {}
`, types.ApplicationVersion, types.ApplicationKind)

	appYaml2 = fmt.Sprintf(`
apiVersion: %s
kind: %s
metadata:
  name: app-2
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
  name: app-1
  namespace: wego-system
resources:
- ./app.yaml
`

	kustomizeYaml2 = `
apiVersion: kustomize.config.k8s.io/v1beta1
commonLabels:
  gitops.weave.works/app-id: 836bac7a-cd97-476b-8989-0d66b5bb82c4
kind: Kustomization
metadata:
  name: app-2
  namespace: wego-system
resources:
- ./app.yaml
`
)

type AppsFixture struct {
	*GomegaWithT
}

func arrangeListOfApps() fstest.MapFS {
	return fstest.MapFS{
		filepath.Join("app-1", types.AppFilename):           {Data: []byte(appYaml)},
		filepath.Join("app-1", types.KustomizationFilename): {Data: []byte(kustomizeYaml)},
		filepath.Join("app-2", types.AppFilename):           {Data: []byte(appYaml2)},
		filepath.Join("app-2", types.KustomizationFilename): {Data: []byte(kustomizeYaml2)},
	}
}

func arrangeSingleApp() fstest.MapFS {
	return fstest.MapFS{
		types.AppFilename:           {Data: []byte(appYaml)},
		types.KustomizationFilename: {Data: []byte(kustomizeYaml)},
	}
}

func setUpAppsTest(t *testing.T) AppsFixture {
	return AppsFixture{
		GomegaWithT: NewGomegaWithT(t),
	}
}

func TestReadApps_MultipleApps(t *testing.T) {
	f := setUpAppsTest(t)

	fileSystem := arrangeListOfApps()
	apps, err := ReadApps(fileSystem, testDir, []string{
		"app-1/kustomization.yaml",
		"app-2/app.yaml",
		"app-2/kustomization.yaml",
		"app-1/app.yaml",
	})

	f.Expect(err).To(BeNil())
	f.Expect(apps).To(HaveLen(2))
	f.Expect(apps["app-1"]).To(Equal(types.App{
		Id:          "836bac7a-cd97-476b-8989-0d66b5bb82c5",
		Name:        "app-1",
		Namespace:   "wego-system",
		Description: "This is a description for the app.",
		DisplayName: "App 1",
		Kustomization: k8types.Kustomization{
			TypeMeta: k8types.TypeMeta{
				APIVersion: k8types.KustomizationVersion,
				Kind:       k8types.KustomizationKind,
			},
			MetaData: &k8types.ObjectMeta{
				Name:      "app-1",
				Namespace: "wego-system",
			},
			CommonLabels: map[string]string{
				types.GitopsLabel("app-id"): "836bac7a-cd97-476b-8989-0d66b5bb82c5",
			},
			Resources: []string{
				"./" + types.AppFilename,
			},
		},
	}))
	f.Expect(apps["app-2"]).To(Equal(types.App{
		Id:          "836bac7a-cd97-476b-8989-0d66b5bb82c4",
		Name:        "app-2",
		Namespace:   "wego-system",
		Description: "This is a description for the app.",
		DisplayName: "",
		Kustomization: k8types.Kustomization{
			TypeMeta: k8types.TypeMeta{
				APIVersion: k8types.KustomizationVersion,
				Kind:       k8types.KustomizationKind,
			},
			MetaData: &k8types.ObjectMeta{
				Name:      "app-2",
				Namespace: "wego-system",
			},
			CommonLabels: map[string]string{
				types.GitopsLabel("app-id"): "836bac7a-cd97-476b-8989-0d66b5bb82c4",
			},
			Resources: []string{
				"./" + types.AppFilename,
			},
		},
	}))
}

func TestReadApps_FileSystemSingleApp(t *testing.T) {
	f := setUpAppsTest(t)

	fileSystem := arrangeSingleApp()
	singleAppDir := filepath.Join(testDir, "app-1")
	apps, err := ReadApps(fileSystem, singleAppDir, []string{
		"kustomization.yaml",
		"app.yaml",
	})

	f.Expect(err).To(BeNil())
	f.Expect(apps).To(HaveLen(1))
	f.Expect(apps["app-1"]).To(Equal(types.App{
		Id:          "836bac7a-cd97-476b-8989-0d66b5bb82c5",
		Name:        "app-1",
		Namespace:   "wego-system",
		Description: "This is a description for the app.",
		DisplayName: "App 1",
		Kustomization: k8types.Kustomization{
			TypeMeta: k8types.TypeMeta{
				APIVersion: k8types.KustomizationVersion,
				Kind:       k8types.KustomizationKind,
			},
			MetaData: &k8types.ObjectMeta{
				Name:      "app-1",
				Namespace: "wego-system",
			},
			CommonLabels: map[string]string{
				types.GitopsLabel("app-id"): "836bac7a-cd97-476b-8989-0d66b5bb82c5",
			},
			Resources: []string{
				"./" + types.AppFilename,
			},
		},
	}))
}
