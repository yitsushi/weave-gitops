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
	AppFilename           = "app.yaml"
	KustomizationFilename = "kustomization.yaml"
)

var (
	AppPathPrefix = fmt.Sprintf("%s/apps", BaseDir)
)

func AppPath(name string) string {
	return fmt.Sprintf("%s/%s", AppPathPrefix, name)
}

func currentPath(fileName string) string {
	return fmt.Sprintf("./%s", fileName)
}

func isKustomizationFile(path string) bool {
	if !strings.HasPrefix(path, AppPathPrefix) {
		return false
	}

	slices := strings.Split(path, "/")
	if len(slices) == 4 && slices[3] == KustomizationFilename {
		return true
	} else {
		return false
	}
}

func appNameFromPath(path string) string {
	if !strings.HasPrefix(path, AppPathPrefix) {
		return ""
	}

	slices := strings.Split(path, "/")
	if len(slices) >= 3 {
		return slices[2]
	} else {
		return ""
	}
}

func GitopsLabel(suffix string) string {
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
			GitopsLabel("app-name"): name,
		},
	}

	return k
}

type App struct {
	Id               string
	Name             string
	Namespace        string
	Description      string
	DisplayName      string
	Kustomization    types.Kustomization
	buckets          map[ObjectKey]v1beta1.Bucket
	kustomizations   map[ObjectKey]v1beta2.Kustomization
	gitRepositories  map[ObjectKey]v1beta1.GitRepository
	helmRepositories map[ObjectKey]v1beta1.HelmRepository
}

func (a *App) path() string {
	return AppPath(a.Name)
}

func (a *App) AddBucketSource(bucket v1beta1.Bucket) ([]repository.File, error) {
	if a.buckets == nil {
		a.buckets = map[ObjectKey]v1beta1.Bucket{}
	}

	objectKey := NewObjectKey(bucket.ObjectMeta)
	a.buckets[objectKey] = bucket

	file, err := bucketSourceFile(a.path(), bucket)
	if err != nil {
		return nil, fmt.Errorf("app.GetFiles bucket: %w", err)
	}

	a.Kustomization.Resources = append(a.Kustomization.Resources, componentFileName(objectKey, v1beta1.BucketKind))
	kustFile, err := kustomizationFile(a.path(), a.Kustomization)
	if err != nil {
		return nil, fmt.Errorf("app.AddBucketSource create bucket file: %w", err)
	}

	return []repository.File{
		file,
		kustFile,
	}, nil
}

func (a *App) AddFluxKustomization(kustomization v1beta2.Kustomization) ([]repository.File, error) {
	if a.kustomizations == nil {
		a.kustomizations = map[ObjectKey]v1beta2.Kustomization{}
	}

	objectKey := NewObjectKey(kustomization.ObjectMeta)
	a.kustomizations[objectKey] = kustomization

	a.Kustomization.Resources = append(a.Kustomization.Resources, componentFileName(objectKey, v1beta2.KustomizationKind))

	file, err := fluxKustomizationFile(a.path(), kustomization)
	if err != nil {
		return nil, fmt.Errorf("app.AddFluxKustomization create flux kustomizaiton file: %w", err)
	}

	kustFile, err := kustomizationFile(a.path(), a.Kustomization)
	if err != nil {
		return nil, fmt.Errorf("app.AddFluxKustomization create kustomizaiton file: %w", err)
	}

	return []repository.File{
		file,
		kustFile,
	}, nil
}

func (a *App) AddGitRepository(gitRepo v1beta1.GitRepository) ([]repository.File, error) {
	if a.gitRepositories == nil {
		a.gitRepositories = map[ObjectKey]v1beta1.GitRepository{}
	}

	objectKey := NewObjectKey(gitRepo.ObjectMeta)
	a.gitRepositories[objectKey] = gitRepo

	a.Kustomization.Resources = append(a.Kustomization.Resources, componentFileName(objectKey, v1beta1.GitRepositoryKind))

	file, err := gitRepositoryFile(a.path(), gitRepo)
	if err != nil {
		return nil, fmt.Errorf("app.AddGitRepository create git repository file: %w", err)
	}

	kustFile, err := kustomizationFile(a.path(), a.Kustomization)
	if err != nil {
		return nil, fmt.Errorf("app.AddGitRepository create kustomizaiton file: %w", err)
	}

	return []repository.File{
		file,
		kustFile,
	}, nil
}

func (a *App) AddHelmRepository(helmRepo v1beta1.HelmRepository) ([]repository.File, error) {
	if a.helmRepositories == nil {
		a.helmRepositories = map[ObjectKey]v1beta1.HelmRepository{}
	}

	objectKey := NewObjectKey(helmRepo.ObjectMeta)
	a.helmRepositories[objectKey] = helmRepo

	a.Kustomization.Resources = append(a.Kustomization.Resources, componentFileName(objectKey, v1beta1.HelmRepositoryKind))

	file, err := helmRepositoryFile(a.path(), helmRepo)
	if err != nil {
		return nil, fmt.Errorf("app.AddHelmRepository create helm repository file: %w", err)
	}

	kustFile, err := kustomizationFile(a.path(), a.Kustomization)
	if err != nil {
		return nil, fmt.Errorf("app.AddHelmRepository create kustomizaiton file: %w", err)
	}

	return []repository.File{
		file,
		kustFile,
	}, nil
}

func (a *App) CustomResource() v1alpha1.Application {
	return v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.ApplicationKind,
			APIVersion: "wego.weave.works/v1alpha1",
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

	appFilePath := filepath.Join(a.path(), AppFilename)

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
		if file, err := fluxKustomizationFile(a.path(), v); err != nil {
			return nil, fmt.Errorf("app files: %w", err)
		} else {
			files = append(files, file)
			paths = append(paths, file.Path)
		}
	}

	if a.Kustomization.MetaData == nil {
		a.Kustomization = NewAppKustomization(a.Name, a.Namespace)
	}

	a.Kustomization.Resources = append(a.Kustomization.Resources, paths...)

	kustomizeData, err := yaml.Marshal(a.Kustomization)
	if err != nil {
		return nil, fmt.Errorf("app %s marshal kustomization into yaml: %w", a.Name, err)
	}

	kustFilePath := filepath.Join(a.path(), KustomizationFilename)

	files = append(files, repository.File{Path: kustFilePath, Data: kustomizeData})

	return files, nil
}
