package app

import (
	"context"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"k8s.io/apimachinery/pkg/types"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . AppService
type AppService interface {
	Add(ctx context.Context, params AddParams) error
	Get(ctx context.Context, name types.NamespacedName) (*wego.Application, error)
	List(ctx context.Context, namespace string) ([]wego.Application, error)
}

type App struct {
	git          git.Git
	flux         flux.Flux
	kube         kube.Kube
	gitProviders gitproviders.GitProviderHandler
	logger       logger.Logger
}

func New(logger logger.Logger, git git.Git, flux flux.Flux, kube kube.Kube, gitProviders gitproviders.GitProviderHandler) *App {
	return &App{
		git:          git,
		flux:         flux,
		kube:         kube,
		gitProviders: gitProviders,
		logger:       logger,
	}
}

// Make sure App implements all the required methods.
var _ AppService = &App{}
