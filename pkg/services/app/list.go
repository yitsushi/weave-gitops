package app

import (
	"context"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
)

func (a *App) List(ctx context.Context, namespace string) ([]wego.Application, error) {
	return a.kube.GetApplications(ctx, namespace)
}
