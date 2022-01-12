package types

import (
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/core/repository"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

const (
	ApplicationKind    = "Application"
	ApplicationVersion = "gitops.weave.works/v1alpha1"

	labelKey              = "gitops.weave.works"
	appFilename           = "app.yaml"
	kustomizationFilename = "kustomization.yaml"
	metadataFilename      = "metadata.json"

	idField          = "id"
	descriptionField = "description"
	versionField     = "version"
)

var (
	appPathPrefix = fmt.Sprintf("%s/apps/", BaseDir)
)

func appPath(name, fileName string) string {
	return fmt.Sprintf("%s/apps/%s/%s", BaseDir, name, fileName)
}

func currentPath(fileName string) string {
	return fmt.Sprintf("./%s", fileName)
}

func isKustomizationFile(path string) bool {
	if !strings.HasPrefix(path, appPathPrefix) {
		return false
	}

	slices := strings.Split(path, "/")
	if len(slices) == 4 && slices[3] == kustomizationFilename {
		return true
	} else {
		return false
	}
}

func appNameFromPath(path string) string {
	if !strings.HasPrefix(path, appPathPrefix) {
		return ""
	}

	slices := strings.Split(path, "/")
	if len(slices) >= 3 {
		return slices[2]
	} else {
		return ""
	}
}

func fileNameFromPath(path string) string {
	if !strings.HasPrefix(path, appPathPrefix) {
		return ""
	}

	slices := strings.Split(path, "/")
	if len(slices) >= 3 {
		return slices[len(slices)-1]
	} else {
		return ""
	}
}

func gitopsLabel(suffix string) string {
	return fmt.Sprintf("%s/%s", labelKey, suffix)
}

type App struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
	DisplayName string `json:"displayName"`
}

func (a App) path(fileName string) string {
	return appPath(a.Name, fileName)
}

func (a App) Kustomization() types.Kustomization {
	k := types.Kustomization{
		TypeMeta: types.TypeMeta{
			Kind:       types.KustomizationKind,
			APIVersion: types.KustomizationVersion,
		},
		MetaData: &types.ObjectMeta{
			Name:      a.Name,
			Namespace: a.Namespace,
		},
		CommonLabels: map[string]string{
			gitopsLabel("app-id"): a.Id,
		},
	}

	return k
}

func (a App) CustomResource() v1alpha1.Application {
	return v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "gitops.weave.works/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.Name,
			Namespace: a.Namespace,
		},
		Spec: v1alpha1.ApplicationSpec{
			Description: a.Description,
			DisplayName: a.DisplayName,
		},
		Status: v1alpha1.ApplicationStatus{},
	}
}

func (a App) Files() ([]repository.File, error) {
	var files []repository.File

	var kustomizeResources []string

	customResource, err := yaml.Marshal(a.CustomResource())
	if err != nil {
		return nil, fmt.Errorf("app %s marshal custom resource into yaml: %w", a.Name, err)
	}

	files = append(files, repository.File{Path: a.path(appFilename), Data: customResource})
	kustomizeResources = append(kustomizeResources, currentPath(appFilename))

	kustomization := a.Kustomization()
	kustomization.Resources = kustomizeResources

	kustomizeData, err := yaml.Marshal(kustomization)
	if err != nil {
		return nil, fmt.Errorf("app %s marshal kustomization into yaml: %w", a.Name, err)
	}

	files = append(files, repository.File{Path: a.path(kustomizationFilename), Data: kustomizeData})

	return files, nil
}
