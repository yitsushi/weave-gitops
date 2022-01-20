package app

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/core/clientset"
	"github.com/weaveworks/weave-gitops/core/gitops/reader"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
	"github.com/weaveworks/weave-gitops/core/source"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type Fetcher interface {
	Get(ctx context.Context, appName, repoName, namespace string) (types.App, error)
	List(ctx context.Context, repoName, namespace string) ([]types.App, error)
}

func NewFetcher(srcService source.Service) Fetcher {
	return &appSourceFetcher{
		srcService: srcService,
	}
}

type appSourceFetcher struct {
	srcService source.Service
}

func (af appSourceFetcher) Get(ctx context.Context, appName, repoName, namespace string) (types.App, error) {
	apps, err := af.getApps(ctx, repoName, namespace)
	if err != nil {
		return types.App{}, fmt.Errorf("appSourceFetcher.Get could not get apps: %w", err)
	}

	if app, ok := apps[appName]; !ok {
		return types.App{}, types.ErrNotFound
	} else {
		return app, nil
	}
}

func (af appSourceFetcher) List(ctx context.Context, repoName, namespace string) ([]types.App, error) {
	appsMap, err := af.getApps(ctx, repoName, namespace)
	if err != nil {
		return nil, fmt.Errorf("appSourceFetcher.List could not get apps: %w", err)
	}

	var apps []types.App
	for _, app := range appsMap {
		apps = append(apps, app)
	}

	return apps, nil
}

func (af appSourceFetcher) getApps(ctx context.Context, repoName, namespace string) (map[string]types.App, error) {
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

type RepoFetcher interface {
	Get(dir string, appName string) (types.App, error)
	List(dir string) ([]types.App, error)
}

func NewRepoFetcher() RepoFetcher {
	return &appRepoFetcher{}
}

type appRepoFetcher struct {
}

func (a appRepoFetcher) Get(dir, appName string) (types.App, error) {
	apps, err := readApps(filepath.Join(dir, types.AppPath(appName)))
	if err != nil {
		return types.App{}, fmt.Errorf("appSourceFetcher.Get could not get apps: %w", err)
	}

	if app, ok := apps[appName]; !ok {
		return types.App{}, types.ErrNotFound
	} else {
		return app, nil
	}
}

func (a appRepoFetcher) List(dir string) ([]types.App, error) {
	appsMap, err := readApps(filepath.Join(dir, types.AppPathPrefix))
	if err != nil {
		return nil, fmt.Errorf("appSourceFetcher.List could not get apps: %w", err)
	}

	var apps []types.App
	for _, app := range appsMap {
		apps = append(apps, app)
	}

	return apps, nil
}

func readApps(dir string) (map[string]types.App, error) {
	var paths []string

	fileSystem := os.DirFS(dir)
	if err := fs.WalkDir(fileSystem, ".", reader.WalkDir(&paths)); err != nil {
		return nil, fmt.Errorf("readApps walking directory: %w", err)
	}

	if appsMap, err := reader.ReadApps(fileSystem, dir, paths); err != nil {
		return nil, fmt.Errorf("readApps reading files: %w", err)
	} else {
		return appsMap, nil
	}
}

type KubeFetcher interface {
	Get(ctx context.Context, client *rest.RESTClient, name, namespace string) (*v1alpha1.Application, error)
	List(ctx context.Context, client *rest.RESTClient, namespace string, opts metav1.ListOptions) (*v1alpha1.ApplicationList, error)
}

func NewKubeAppFetcher() KubeFetcher {
	return &appKubeFetcher{}
}

type appKubeFetcher struct {
}

func (a appKubeFetcher) Get(ctx context.Context, client *rest.RESTClient, name, namespace string) (result *v1alpha1.Application, err error) {
	result = &v1alpha1.Application{}
	err = client.Get().
		Namespace(namespace).
		Resource(apps).
		Name(name).
		Do(ctx).
		Into(result)

	return
}

func (a appKubeFetcher) List(ctx context.Context, client *rest.RESTClient, namespace string, opts metav1.ListOptions) (result *v1alpha1.ApplicationList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	result = &v1alpha1.ApplicationList{}
	err = client.Get().
		Namespace(namespace).
		Resource(apps).
		VersionedParams(&opts, clientset.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)

	return
}
