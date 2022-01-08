package types

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/core/source"
	"sigs.k8s.io/kustomize/api/types"
)

func FileJsonToRepo(objects []source.FileJson) Repo {
	return Repo{}
}

func FileJsonToApps(files []source.FileJson) (map[string]App, error) {
	apps := map[string]App{}

	for _, file := range files {
		appName := appNameFromPath(file.Path)
		if appName == "" {
			continue
		}

		if _, ok := apps[appName]; !ok {
			apps[appName] = App{
				Name: appName,
			}
		}

		app := apps[appName]

		isAppFile := strings.HasSuffix(file.Path, appFilename)
		if isAppFile {
			var appResource v1alpha1.Application
			err := mapstructure.Decode(file.Data, &appResource)
			if err != nil {
				return nil, fmt.Errorf("could not decode kustomization file into struct: %w", err)
			}

			app.Description = appResource.Spec.Description
		} else if isKustomizationFile(file.Path) {
			var kustomization types.Kustomization
			err := mapstructure.Decode(file.Data, &kustomization)
			if err != nil {
				return nil, fmt.Errorf("could not decode kustomization file into struct: %w", err)
			}

			app.Namespace = kustomization.MetaData.Namespace
			app.Id = kustomization.CommonLabels[gitopsLabel("app-id")]
		}

		apps[appName] = app
	}

	return apps, nil
}
