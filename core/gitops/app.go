package gitops

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

const (
	labelKey              = "weave.works.gitops"
	kustomizationFilename = "kustomization.yaml"
)

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
	return fmt.Sprintf("/apps/%s/%s", a.Name, fileName)
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
			gitopsLabel("app"): a.Name,
		},
	}

	return k
}

func (a App) Files() ([]File, error) {
	var files []File

	kustomizeData, err := yaml.Marshal(a.Kustomization())
	if err != nil {
		return nil, fmt.Errorf("app %s marshal kustomization into yaml: %w", a.Name, err)
	}

	files = append(files, File{Path: a.path(kustomizationFilename), Data: kustomizeData})

	metadata := map[string]interface{}{
		"id":          a.Id,
		"description": a.Description,
		"version":     1,
	}
	metadataData, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("app %s marshal metadata into json: %w", a.Name, err)
	}

	files = append(files, File{Path: a.path("metadata.json"), Data: metadataData})

	return files, nil
}
