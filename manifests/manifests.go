package manifests

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/weaveworks/weave-gitops/core/repository"
)

const (
	gitopsAppManifestDir     = "gitops/app"
	gitopsRuntimeManifestDir = "gitops/runtime"
	wegoManifestsDir         = "wego-app"
	templateExtensionLen     = 4
)

var (
	//go:embed gitops/runtime/wego.weave.works_apps.yaml
	AppCRD []byte
	//go:embed wego-app/*
	wegoAppTemplates embed.FS
	//go:embed gitops/app/*
	gitopsAppTemplates embed.FS
	//go:embed gitops/runtime/*
	gitopsRuntimeTemplates embed.FS
)

type Params struct {
	AppVersion string
	Namespace  string
}

// GitopsManifests generates manifests for Weave GitOps's application and runtime
func GitopsManifests(pathPrefix string, params Params) ([]repository.File, error) {
	appFiles, err := getManifestFiles(pathPrefix, gitopsAppTemplates, gitopsAppManifestDir, params)
	if err != nil {
		return nil, fmt.Errorf("failed to read gitops app templates: %w", err)
	}

	runtimeFiles, err := getManifestFiles(pathPrefix, gitopsRuntimeTemplates, gitopsRuntimeManifestDir, params)
	if err != nil {
		return nil, fmt.Errorf("failed to read gitops runtime templates: %w", err)
	}

	return append(appFiles, runtimeFiles...), nil
}

func getManifestFiles(pathPrefix string, embedded embed.FS, dir string, params Params) ([]repository.File, error) {
	templates, err := fs.ReadDir(embedded, dir)
	if err != nil {
		return nil, fmt.Errorf("failed reading templates directory: %w", err)
	}

	var files []repository.File

	for _, template := range templates {
		tplName := template.Name()
		filePath := filepath.Join(dir, tplName)

		data, err := fs.ReadFile(embedded, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed reading template %s: %w", tplName, err)
		}

		var file repository.File
		if !strings.HasSuffix(filePath, ".tpl") {
			file = repository.File{
				Path: filepath.Join(pathPrefix, filePath),
				Data: data,
			}
		} else {
			manifest, err := executeTemplate(tplName, string(data), params)
			if err != nil {
				return nil, fmt.Errorf("failed executing template: %s: %w", tplName, err)
			}

			file = repository.File{
				Path: filepath.Join(pathPrefix, filePath[:len(filePath)-templateExtensionLen]),
				Data: manifest,
			}
		}

		files = append(files, file)
	}

	return files, nil
}

// GenerateManifests generates weave-gitops manifests from a template
func GenerateManifests(params Params) ([][]byte, error) {
	templates, err := fs.ReadDir(wegoAppTemplates, wegoManifestsDir)
	if err != nil {
		return nil, fmt.Errorf("failed reading templates directory: %w", err)
	}

	var manifests [][]byte

	for _, template := range templates {
		tplName := template.Name()

		data, err := fs.ReadFile(wegoAppTemplates, filepath.Join(wegoManifestsDir, tplName))
		if err != nil {
			return nil, fmt.Errorf("failed reading template %s: %w", tplName, err)
		}

		manifest, err := executeTemplate(tplName, string(data), params)
		if err != nil {
			return nil, fmt.Errorf("failed executing template: %s: %w", tplName, err)
		}

		manifests = append(manifests, manifest)
	}

	return manifests, nil
}

func executeTemplate(name string, tplData string, params Params) ([]byte, error) {
	template, err := template.New(name).Parse(tplData)
	if err != nil {
		return nil, fmt.Errorf("error parsing template %s: %w", name, err)
	}

	yaml := &bytes.Buffer{}

	err = template.Execute(yaml, params)
	if err != nil {
		return nil, fmt.Errorf("error injecting values to template: %w", err)
	}

	return yaml.Bytes(), nil
}
