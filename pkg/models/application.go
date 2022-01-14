package models

import (
	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
)

type AutomationType string

type SourceType string

const (
	AutomationTypeHelm      AutomationType = "helm"
	AutomationTypeKustomize AutomationType = "kustomize"

	SourceTypeGit  SourceType = "git"
	SourceTypeHelm SourceType = "helm"
)

type Application struct {
	Name                string
	Namespace           string
	HelmSourceURL       string
	GitSourceURL        gitproviders.RepoURL
	ConfigRepo          gitproviders.RepoURL
	Branch              string
	Path                string
	AutomationType      AutomationType
	SourceType          SourceType
	HelmTargetNamespace string
	RepoVisibility      gitprovider.RepositoryVisibility
}

func IsExternalConfigRepo(url string) bool {
	return url != ""
}
