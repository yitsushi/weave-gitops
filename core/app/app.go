package app

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/types"
)

const (
	labelKey = "weave.works.gitops"
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
