package types

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/source"
	"sigs.k8s.io/kustomize/api/types"
)

type conversionFixture struct {
	*GomegaWithT
}

func setUpConversionTest(t *testing.T) appFixture {
	return appFixture{
		GomegaWithT: NewGomegaWithT(t),
	}
}

func TestConvertApp(t *testing.T) {
	f := setUpConversionTest(t)
	applicationFile := source.FileJson{
		Path: appPath("app-1", appFilename),
		Data: map[string]interface{}{
			"kind":       ApplicationKind,
			"apiVersion": ApplicationVersion,
			"metadata": map[string]interface{}{
				"name":      "app-1",
				"namespace": testNamespace,
			},
			"spec": map[string]interface{}{
				"description": "This is a test app",
			},
		},
	}

	kustomization := map[string]interface{}{
		"kind":       types.KustomizationKind,
		"apiVersion": types.KustomizationVersion,
		"metadata": map[string]interface{}{
			"name":      "app-1",
			"namespace": testNamespace,
		},
		"commonLabels": map[string]string{
			gitopsLabel("app-id"): "12345",
		},
	}

	kustomizationFile := source.FileJson{
		Path: appPath("app-1", kustomizationFilename),
		Data: kustomization,
	}

	var files = []source.FileJson{
		applicationFile,
		kustomizationFile,
	}

	apps, err := FileJsonToApps(files)

	f.Expect(err).To(BeNil())
	f.Expect(apps).To(HaveLen(1))
	f.Expect(apps["app-1"]).To(Equal(App{
		Id:          "12345",
		Name:        "app-1",
		Namespace:   testNamespace,
		Description: "This is a test app",
	}))
}
