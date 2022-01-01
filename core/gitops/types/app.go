package types

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops/core/repository"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

const (
	labelKey              = "app.weave.works.gitops"
	kustomizationFilename = "kustomization.yaml"
	metadataFilename      = "metadata.json"

	idField          = "id"
	descriptionField = "description"
	versionField     = "version"
)

var (
	appPathPrefix = fmt.Sprintf("%s/apps/", baseDir)
)

func appPath(name, fileName string) string {
	return fmt.Sprintf("%s/apps/%s/%s", baseDir, name, fileName)
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
			Annotations: map[string]string{
				gitopsLabel("app-id"):          a.Id,
				gitopsLabel("app-description"): a.Description,
				gitopsLabel("app-version"):     "v1beta1",
			},
		},
		CommonLabels: map[string]string{
			gitopsLabel("app"): a.Name,
		},
	}

	return k
}

func (a App) Files() ([]repository.File, error) {
	var files []repository.File

	kustomizeData, err := yaml.Marshal(a.Kustomization())
	if err != nil {
		return nil, fmt.Errorf("app %s marshal kustomization into yaml: %w", a.Name, err)
	}

	files = append(files, repository.File{Path: a.path(kustomizationFilename), Data: kustomizeData})

	metadata := map[string]interface{}{
		idField:          a.Id,
		descriptionField: a.Description,
		versionField:     1,
	}

	metadataData, err := json.MarshalIndent(metadata, "", "\t")
	if err != nil {
		return nil, fmt.Errorf("app %s marshal metadata into json: %w", a.Name, err)
	}

	files = append(files, repository.File{Path: a.path(metadataFilename), Data: metadataData})

	return files, nil
}
