package app

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/core/gitops/types"
	"github.com/weaveworks/weave-gitops/core/source"
)

type Fetcher interface {
	Get(ctx context.Context, appName, repoName, namespace string) (types.App, error)
	List(ctx context.Context, repoName, namespace string) ([]types.App, error)
}

func NewFetcher(srcService source.Service) Fetcher {
	return &appFetcher{
		srcService: srcService,
	}
}

type appFetcher struct {
	srcService source.Service
}

func (af appFetcher) Get(ctx context.Context, appName, repoName, namespace string) (types.App, error) {
	apps, err := af.getApps(ctx, repoName, namespace)
	if err != nil {
		return types.App{}, fmt.Errorf("appFetcher.Get could not get apps: %w", err)
	}

	if app, ok := apps[appName]; !ok {
		return types.App{}, types.ErrNotFound
	} else {
		return app, nil
	}
}

func (af appFetcher) List(ctx context.Context, repoName, namespace string) ([]types.App, error) {
	appsMap, err := af.getApps(ctx, repoName, namespace)
	if err != nil {
		return nil, fmt.Errorf("appFetcher.List could not get apps: %w", err)
	}

	var apps []types.App
	for _, app := range appsMap {
		apps = append(apps, app)
	}

	return apps, nil
}

func (af appFetcher) getApps(ctx context.Context, repoName, namespace string) (map[string]types.App, error) {
	files, err := af.srcService.GetArtifact(ctx, repoName, namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch artifacts from %s at %s: %w", repoName, namespace, err)
	}

	appsMap, err := types.FileJsonToApps(files)
	if err != nil {
		return nil, fmt.Errorf("unable to convert JSON files into apps: %w", err)
	}

	return appsMap, nil
}
