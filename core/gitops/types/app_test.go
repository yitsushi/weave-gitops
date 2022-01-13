package types

import (
	"testing"

	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
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

	appName := appNameFromPath(BaseDir)
	f.Expect(appName).To(Equal(""))

	appName = appNameFromPath(BaseDir + "/apps_r_us/my-app/test.yaml")
	f.Expect(appName).To(Equal(""), "app name should be an empty string with an invalid path")
}

func TestAppNameFromPath_AnAppPathNoAppSubdirectory(t *testing.T) {
	f := setUpAppTest(t)
	appName := appNameFromPath(BaseDir + "/apps")

	f.Expect(appName).To(Equal(""))
}

func TestAppNameFromPath_ValidPaths(t *testing.T) {
	f := setUpAppTest(t)
	appName := appNameFromPath(BaseDir + "/apps/my-app-1")
	f.Expect(appName).To(Equal("my-app-1"))

	appName = appNameFromPath(BaseDir + "/apps/my-app-2/test.yaml")
	f.Expect(appName).To(Equal("my-app-2"))
}

func TestApp_AddAndGetKustomization(t *testing.T) {
	f := setUpAppTest(t)

	app := App{
		Id:          "12345",
		Name:        "my-app",
		Namespace:   testNamespace,
		Description: "This is my test application, it's going to take over the world",
	}

	kust1 := v1beta2.Kustomization{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1beta2.KustomizationKind,
			APIVersion: v1beta2.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kust-1",
			Namespace: testNamespace,
		},
		Spec:   v1beta2.KustomizationSpec{},
		Status: v1beta2.KustomizationStatus{},
	}
	app.AddFluxKustomization(kust1)

	kust2 := kust1
	kust2.ObjectMeta.Namespace = "extra-namespace"
	app.AddFluxKustomization(kust2)

	f.Expect(app.kustomizations).To(HaveLen(2))

	k, ok := app.GetFluxKustomization(ObjectKey{Name: "kust-1", Namespace: "extra-namespace"})
	f.Expect(ok).To(BeTrue())
	f.Expect(k).To(Equal(kust2))

	k3, ok := app.GetFluxKustomization(ObjectKey{Name: "fake-kust", Namespace: "bad-robot"})
	f.Expect(ok).To(BeFalse())
	f.Expect(k3).To(Equal(v1beta2.Kustomization{}))
}

func TestApp_KustomizationFiles(t *testing.T) {

}

func TestAppFiles_Success(t *testing.T) {
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

	expectedKustomize := NewAppKustomization("my-app", testNamespace)
	expectedKustomize.Resources = []string{files[0].Path}
	expectedData, _ := yaml.Marshal(expectedKustomize)
	f.Expect(files[1].Data).To(Equal(expectedData))

}
