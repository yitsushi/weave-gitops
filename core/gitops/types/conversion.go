package types

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
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

		isMetadataFile := strings.HasSuffix(file.Path, metadataFilename)
		if isMetadataFile {
			app.Id = file.Data[idField].(string)
			app.Description = file.Data[descriptionField].(string)
		} else if isKustomizationFile(file.Path) {
			var kustomization types.Kustomization
			err := mapstructure.Decode(file.Data, &kustomization)
			if err != nil {
				return nil, fmt.Errorf("could not decode kustomization file into struct: %w", err)
			}

			app.Namespace = kustomization.MetaData.Namespace
		}

		apps[appName] = app
	}

	return apps, nil
}
