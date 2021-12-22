package types

import (
	"fmt"
	"path/filepath"

	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/core/repository"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

func fluxKustomizationFile(prefixPath string, k v1beta2.Kustomization) (repository.File, error) {
	fileName := componentFilePath(prefixPath, NewObjectKey(k.ObjectMeta), v1beta2.KustomizationKind)
	filePath := filepath.Join(prefixPath, fileName)

	data, err := yaml.Marshal(k)
	if err != nil {
		return repository.File{}, fmt.Errorf("unable to marshal flux kustomization %s/%s: %w", k.ObjectMeta.Name, k.ObjectMeta.Namespace, err)
	}

	return repository.File{Path: filePath, Data: data}, nil
}

func kustomizationFile(prefixPath string, k types.Kustomization) (repository.File, error) {
	filePath := filepath.Join(prefixPath, KustomizationFilename)

	data, err := yaml.Marshal(k)
	if err != nil {
		return repository.File{}, fmt.Errorf("unable to marshal kustomization: %w", err)
	}

	return repository.File{Path: filePath, Data: data}, nil
}
