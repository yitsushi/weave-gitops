package reader

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
)

type ErrInvalidPath struct {
	path string
}

func (e ErrInvalidPath) Error() string {
	return fmt.Sprintf("invalid path for file %s", e.path)
}

func appNameFromPath(dir, path string) string {
	if strings.HasSuffix(dir, types.AppPathPrefix) {
		slices := strings.Split(path, "/")

		if len(slices) > 0 {
			return slices[0]
		}
	} else {
		slices := strings.Split(dir, "/")
		if len(slices) > 0 {
			return slices[len(slices)-1]
		}
	}

	return ""
}

func parentAppFolder(dir, path string) bool {
	slices := strings.Split(path, "/")
	if strings.HasSuffix(dir, types.AppPathPrefix) {
		return len(slices) == 2
	} else {
		return len(slices) == 1
	}
}

func isKustomizationFile(dir, path string) bool {
	slices := strings.Split(path, "/")
	if strings.HasSuffix(dir, types.AppPathPrefix) {
		return len(slices) == 2 && slices[1] == types.KustomizationFilename
	} else {
		return len(slices) == 1 && slices[0] == types.KustomizationFilename
	}
}

type App interface {
	Read(paths []string) (map[string]types.App, error)
}

func ReadApps(fileSystem fs.FS, dir string, paths []string) (map[string]types.App, error) {
	apps := map[string]types.App{}

	for _, path := range paths {
		appName := appNameFromPath(dir, path)
		if appName == "" {
			continue
		}

		if _, ok := apps[appName]; !ok {
			apps[appName] = types.App{
				Name: appName,
			}
		}

		app := apps[appName]

		isAppFile := strings.HasSuffix(path, types.AppFilename)
		data, err := readJsonOrYamlFile(fileSystem, path)
		if err != nil {
			return nil, fmt.Errorf("read apps error reading file %s: %w", path, err)
		}

		if isAppFile {
			var appResource v1alpha1.Application
			err := mapstructure.Decode(data, &appResource)
			if err != nil {
				return nil, fmt.Errorf("could not decode kustomization file into struct: %w", err)
			}

			app.Description = appResource.Spec.Description
			app.DisplayName = appResource.Spec.DisplayName
		} else if isKustomizationFile(dir, path) {
			kustomization, err := readKustomizationFile(fileSystem, path)
			if err != nil {
				return nil, fmt.Errorf("ReadApps read kustomization file: %w", err)
			}

			app.Namespace = kustomization.MetaData.Namespace
			app.Id = kustomization.CommonLabels[types.GitopsLabel("app-id")]
			app.Kustomization = kustomization
		}

		apps[appName] = app
	}

	return apps, nil
}
