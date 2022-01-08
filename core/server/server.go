package server

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/core/gitops/app"
	"github.com/weaveworks/weave-gitops/core/source"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

func Hydrate(ctx context.Context, mux *runtime.ServeMux) error {
	k8sClient, err := kube.NewClient()
	if err != nil {
		return err
	}

	sourceSvc := source.NewService(k8sClient, source.GitopsRuntimeExclusionList)
	appFetcher := app.NewFetcher(sourceSvc)

	newAppServer := NewAppServer(appFetcher)
	if err := pb.RegisterAppsHandlerServer(ctx, mux, newAppServer); err != nil {
		return fmt.Errorf("could not register new app: %w", err)
	}

	return nil
}
