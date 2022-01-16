package reader

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
	k8types "sigs.k8s.io/kustomize/api/types"
)

type ErrInvalidPath struct {
	path string
}

func (e ErrInvalidPath) Error() string {
	return fmt.Sprintf("invalid path for file %s", e.path)
}

func appNameFromPath(path string) string {
	if !strings.HasPrefix(path, types.AppPathPrefix) {
		return ""
	}

	slices := strings.Split(path, "/")
	if len(slices) >= 3 {
		return slices[2]
	} else {
		return ""
	}
}

func isKustomizationFile(path string) bool {
	if !strings.HasPrefix(path, types.AppPathPrefix) {
		return false
	}

	slices := strings.Split(path, "/")
	if len(slices) == 4 && slices[3] == types.KustomizationFilename {
		return true
	} else {
		return false
	}
}

type App interface {
	Read(paths []string) (map[string]types.App, error)
}

func ReadApps(fileSystem fs.FS, paths []string) (map[string]types.App, error) {
	apps := map[string]types.App{}

	for _, path := range paths {

		if !strings.HasPrefix(path, types.AppPathPrefix) {
			return nil, ErrInvalidPath{path: path}
		}

		appName := appNameFromPath(path)
		if appName == "" {
			continue
		}

		if _, ok := apps[appName]; !ok {
			apps[appName] = types.App{
				Name: appName,
			}
		}

		app := apps[appName]

		data, err := readJsonOrYamlFile(fileSystem, path)
		if err != nil {
			return nil, fmt.Errorf("read apps error reading file %s: %w", path, err)
		}

		isAppFile := strings.HasSuffix(path, types.AppFilename)
		if isAppFile {
			var appResource v1alpha1.Application
			err := mapstructure.Decode(data, &appResource)
			if err != nil {
				return nil, fmt.Errorf("could not decode kustomization file into struct: %w", err)
			}

			app.Description = appResource.Spec.Description
		} else if isKustomizationFile(path) {
			var kustomization k8types.Kustomization
			err := mapstructure.Decode(data, &kustomization)
			if err != nil {
				return nil, fmt.Errorf("could not decode kustomization file into struct: %w", err)
			}

			app.Namespace = kustomization.MetaData.Namespace
			app.Id = kustomization.CommonLabels[types.GitopsLabel("app-id")]
		}

		apps[appName] = app
	}

	return apps, nil
}
