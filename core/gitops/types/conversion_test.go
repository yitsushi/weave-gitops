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
	metadata := map[string]interface{}{
		idField:          "12345",
		descriptionField: "This is a test app",
		versionField:     1,
	}

	metadataFile := source.FileJson{
		Path: appPath("app-1", metadataFilename),
		Data: metadata,
	}

	kustomization := map[string]interface{}{
		"kind":       types.KustomizationKind,
		"apiVersion": types.KustomizationVersion,
		"metadata": map[string]interface{}{
			"name":      "app-1",
			"namespace": testNamespace,
			"annotations": map[string]string{
				gitopsLabel("app-id"):          "12345",
				gitopsLabel("app-description"): "This is a test app",
				gitopsLabel("app-version"):     "v1beta1",
			},
		},
		"commonLabels": map[string]string{
			gitopsLabel("app"): "app-1",
		},
	}

	kustomizationFile := source.FileJson{
		Path: appPath("app-1", kustomizationFilename),
		Data: kustomization,
	}

	var files = []source.FileJson{
		kustomizationFile,
		metadataFile,
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
