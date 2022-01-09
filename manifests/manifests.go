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
	gitopsManifestDir = "gitops"
	wegoManifestsDir  = "wego-app"
)

var (
	//go:embed crds/wego.weave.works_apps.yaml
	AppCRD []byte
	//go:embed wego-app/*
	wegoAppTemplates embed.FS
	//go:embed gitops/*
	gitopsAppTemplates embed.FS
)

type Params struct {
	AppVersion string
	Namespace  string
}

// GitopsManifests generates manifests for Weave GitOps's application and runtime
func GitopsManifests(params Params) ([]repository.File, error) {
	templates, err := fs.ReadDir(gitopsAppTemplates, gitopsManifestDir)
	if err != nil {
		return nil, fmt.Errorf("failed reading templates directory: %w", err)
	}

	var manifests []repository.File

	for _, template := range templates {
		tplName := template.Name()

		filePath := filepath.Join(gitopsManifestDir, tplName)
		templateBytes, err := fs.ReadFile(gitopsAppTemplates, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed reading template %s: %w", tplName, err)
		}

		data, err := executeTemplate(tplName, string(templateBytes), params)
		if err != nil {
			return nil, fmt.Errorf("failed executing template: %s: %w", tplName, err)
		}

		if strings.HasSuffix(filePath, ".tpl") {

		}

		manifests = append(manifests, repository.File{
			Path: filePath,
			Data: data,
		})
	}

	return manifests, nil
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
