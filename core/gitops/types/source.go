package types

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/weaveworks/weave-gitops/core/repository"
	"sigs.k8s.io/yaml"
)

func componentFileName(ok ObjectKey, kind string) string {
	return fmt.Sprintf("%s-%s-%s.yaml", ok.Name, ok.Namespace, strings.ToLower(kind))
}

func componentFilePath(prefixPath string, ok ObjectKey, kind string) string {
	return filepath.Join(prefixPath, componentFileName(ok, kind))
}

func bucketSourceFile(prefixPath string, b v1beta1.Bucket) (repository.File, error) {
	filePath := componentFilePath(prefixPath, NewObjectKey(b.ObjectMeta), v1beta1.BucketKind)
	data, err := yaml.Marshal(b)
	if err != nil {
		return repository.File{}, fmt.Errorf("unable to marshal bucket source %s/%s: %w", b.ObjectMeta.Name, b.ObjectMeta.Namespace, err)
	}

	return repository.File{Path: filePath, Data: data}, nil
}

func gitRepositoryFile(prefixPath string, gr v1beta1.GitRepository) (repository.File, error) {
	filePath := componentFilePath(prefixPath, NewObjectKey(gr.ObjectMeta), v1beta1.GitRepositoryKind)
	data, err := yaml.Marshal(gr)
	if err != nil {
		return repository.File{}, fmt.Errorf("unable to marshal git repository %s/%s: %w", gr.ObjectMeta.Name, gr.ObjectMeta.Namespace, err)
	}

	return repository.File{Path: filePath, Data: data}, nil
}

func helmRepositoryFile(prefixPath string, hr v1beta1.HelmRepository) (repository.File, error) {
	filePath := componentFilePath(prefixPath, NewObjectKey(hr.ObjectMeta), v1beta1.HelmRepositoryKind)

	data, err := yaml.Marshal(hr)
	if err != nil {
		return repository.File{}, fmt.Errorf("unable to marshal helm repository %s/%s: %w", hr.ObjectMeta.Name, hr.ObjectMeta.Namespace, err)
	}

	return repository.File{Path: filePath, Data: data}, nil
}
