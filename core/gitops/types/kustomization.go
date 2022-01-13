package types

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/core/repository"
	"sigs.k8s.io/yaml"
)

func kustomizationFile(prefixPath string, k v1beta2.Kustomization) (repository.File, error) {
	fileName := fmt.Sprintf("%s-%s-%s.yaml", k.ObjectMeta.Name, k.ObjectMeta.Namespace, strings.ToLower(v1beta2.KustomizationKind))
	filePath := filepath.Join(prefixPath, fileName)

	data, err := yaml.Marshal(k)
	if err != nil {
		return repository.File{}, fmt.Errorf("unable to marshal kustomization %s/%s: %w", k.ObjectMeta.Name, k.ObjectMeta.Namespace, err)
	}

	return repository.File{Path: filePath, Data: data}, nil
}
