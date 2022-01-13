package server

import (
	"context"
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/core/gitops/app"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
	"github.com/weaveworks/weave-gitops/core/repository"
	"github.com/weaveworks/weave-gitops/core/source"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Hydrate(ctx context.Context, mux *runtime.ServeMux) error {
	k8sClient, err := kube.NewClient()
	if err != nil {
		return err
	}

	sourceSvc := source.NewService(k8sClient, source.GitopsRuntimeExclusionList)

	repoManager := repository.NewRepoManager()

	writerSvc := repository.NewGitWriter(true)
	appCreator := app.NewCreator(writerSvc)
	appFetcher := app.NewFetcher(sourceSvc)

	deleterSvc := repository.NewGitDeleter(true)
	appRemover := app.NewRemover(deleterSvc, appFetcher)

	newAppServer := NewAppServer(appCreator, appFetcher, appRemover, sourceSvc, repoManager)
	if err := pb.RegisterAppsHandlerServer(ctx, mux, newAppServer); err != nil {
		return fmt.Errorf("could not register new app: %w", err)
	}

	return nil
}

func getRepo(sourceSvc source.Service, manager repository.Manager, repoName string) (*git.Repository, *ssh.PublicKeys, error) {
	sourceRepo, err := sourceSvc.Get(context.Background(), repoName, types.FluxNamespace)
	if err != nil {
		return nil, nil, fmt.Errorf("getRepo: unable to get config repo: %s", err.Error())
	}

	key, err := sourceSvc.GetClientKey(context.Background(), types.FluxNamespace, sourceRepo)
	if err != nil {
		return nil, nil, fmt.Errorf("getRepo: unable to get git repo key: %s", err.Error())
	}

	repo, err := manager.Get(context.Background(), key, sourceRepo.Spec.URL, "test")
	if err != nil {
		return nil, nil, fmt.Errorf("getRepo: unable to get git repo: %s", err.Error())
	}

	return repo, key, nil
}

func intervalDuration(input *pb.Interval) metav1.Duration {
	return metav1.Duration{Duration: time.Duration(input.Hours)*time.Hour + time.Duration(input.Minutes)*time.Minute + time.Duration(input.Seconds)*time.Second}
}
