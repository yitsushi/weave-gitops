package types

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/source-controller/api/v1beta1"
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

func NewAppKustomization(name, namespace string) types.Kustomization {
	k := types.Kustomization{
		TypeMeta: types.TypeMeta{
			Kind:       types.KustomizationKind,
			APIVersion: types.KustomizationVersion,
		},
		MetaData: &types.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		CommonLabels: map[string]string{
			gitopsLabel("app-name"): name,
		},
	}

	return k
}

type App struct {
	Id              string                              `json:"id"`
	Name            string                              `json:"name"`
	Namespace       string                              `json:"namespace"`
	Description     string                              `json:"description"`
	DisplayName     string                              `json:"displayName"`
	kustomization   types.Kustomization                 `json:"kustomization"`
	kustomizations  map[ObjectKey]v1beta2.Kustomization `json:"kustomizations"`
	gitRepositories map[ObjectKey]v1beta1.GitRepository `json:"gitRepositories"`
}

func (a *App) path() string {
	return fmt.Sprintf("%s/apps/%s", BaseDir, a.Name)
}

func (a *App) AddFluxKustomization(kustomization v1beta2.Kustomization) {
	if a.kustomizations == nil {
		a.kustomizations = map[ObjectKey]v1beta2.Kustomization{}
	}
	a.kustomizations[NewObjectKey(kustomization.ObjectMeta)] = kustomization
}

func (a *App) GetFluxKustomization(key ObjectKey) (v1beta2.Kustomization, bool) {
	k, ok := a.kustomizations[key]
	return k, ok
}

func (a *App) AddGitRepository(gitRepo v1beta1.GitRepository) {
	if a.gitRepositories == nil {
		a.gitRepositories = map[ObjectKey]v1beta1.GitRepository{}
	}
	a.gitRepositories[NewObjectKey(gitRepo.ObjectMeta)] = gitRepo
}

func (a *App) GetGitRepository(key ObjectKey) (v1beta1.GitRepository, bool) {
	gr, ok := a.gitRepositories[key]
	return gr, ok
}

func (a *App) CustomResource() v1alpha1.Application {
	return v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       ApplicationKind,
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

func (a *App) Files() ([]repository.File, error) {
	var files []repository.File

	var paths []string

	customResource, err := yaml.Marshal(a.CustomResource())
	if err != nil {
		return nil, fmt.Errorf("app %s marshal custom resource into yaml: %w", a.Name, err)
	}

	appFilePath := filepath.Join(a.path(), appFilename)
	files = append(files, repository.File{Path: appFilePath, Data: customResource})
	paths = append(paths, currentPath(appFilePath))

	for _, v := range a.gitRepositories {
		if file, err := gitRepositoryFile(a.path(), v); err != nil {
			return nil, fmt.Errorf("app files: %w", err)
		} else {
			files = append(files, file)
			paths = append(paths, file.Path)
		}
	}

	for _, v := range a.kustomizations {
		if file, err := kustomizationFile(a.path(), v); err != nil {
			return nil, fmt.Errorf("app files: %w", err)
		} else {
			files = append(files, file)
			paths = append(paths, file.Path)
		}
	}

	if a.kustomization.MetaData == nil {
		a.kustomization = NewAppKustomization(a.Name, a.Namespace)
	}

	a.kustomization.Resources = append(a.kustomization.Resources, paths...)

	kustomizeData, err := yaml.Marshal(a.kustomization)
	if err != nil {
		return nil, fmt.Errorf("app %s marshal kustomization into yaml: %w", a.Name, err)
	}

	kustFilePath := filepath.Join(a.path(), kustomizationFilename)
	files = append(files, repository.File{Path: kustFilePath, Data: kustomizeData})

	return files, nil
}
