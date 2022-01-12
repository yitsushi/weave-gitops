package server

import (
	"context"

	"github.com/weaveworks/weave-gitops/core/gitops/app"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
	"github.com/weaveworks/weave-gitops/core/repository"
	"github.com/weaveworks/weave-gitops/core/source"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newProtoApp(a types.App) *pb.App {
	return &pb.App{
		Id:          a.Id,
		Name:        a.Name,
		Description: a.Description,
	}
}

type appServer struct {
	pb.UnimplementedAppsServer

	creator     app.Creator
	fetcher     app.Fetcher
	remover     app.Remover
	repoManager repository.Manager
	sourceSvc   source.Service
}

func NewAppServer(creator app.Creator, fetcher app.Fetcher, remover app.Remover, sourceSvc source.Service, repoManager repository.Manager) pb.AppsServer {
	return &appServer{
		creator:     creator,
		fetcher:     fetcher,
		remover:     remover,
		repoManager: repoManager,
		sourceSvc:   sourceSvc,
	}
}

func (a *appServer) AddApp(_ context.Context, msg *pb.AddAppRequest) (*pb.AddAppResponse, error) {
	sourceRepo, err := a.sourceSvc.Get(context.Background(), msg.RepoName, types.FluxNamespace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create app: unable to get config repo: %s", err.Error())
	}

	key, err := a.sourceSvc.GetClientKey(context.Background(), types.FluxNamespace, sourceRepo)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create app: unable to get git repo key: %s", err.Error())
	}

	repo, err := a.repoManager.Get(context.Background(), key, sourceRepo.Spec.URL, "test")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create app: unable to get git repo: %s", err.Error())
	}

	app, err := a.creator.Create(repo, key, app.CreateInput{
		Name:        msg.Name,
		Namespace:   types.FluxNamespace,
		Description: msg.Description,
		DisplayName: msg.DisplayName,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to create new app: %s", err.Error())
	}

	return &pb.AddAppResponse{
		App:     newProtoApp(app),
		Success: true,
	}, nil
}

func (a *appServer) GetApp(_ context.Context, msg *pb.GetAppRequest) (*pb.GetAppResponse, error) {
	app, err := a.fetcher.Get(context.Background(), msg.AppName, msg.RepoName, types.FluxNamespace)
	if err == types.ErrNotFound {
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetAppResponse{App: newProtoApp(app)}, nil
}

func (a *appServer) ListApps(_ context.Context, msg *pb.ListAppRequest) (*pb.ListAppResponse, error) {
	apps, err := a.fetcher.List(context.Background(), msg.RepoName, types.FluxNamespace)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var protoApps []*pb.App
	for _, a := range apps {
		protoApps = append(protoApps, newProtoApp(a))
	}

	return &pb.ListAppResponse{
		Apps: protoApps,
	}, nil
}

func (a *appServer) RemoveApp(ctx context.Context, msg *pb.RemoveAppRequest) (*pb.RemoveAppResponse, error) {
	sourceRepo, err := a.sourceSvc.Get(context.Background(), msg.RepoName, types.FluxNamespace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "remove app: unable to get config repo: %s", err.Error())
	}

	key, err := a.sourceSvc.GetClientKey(context.Background(), types.FluxNamespace, sourceRepo)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "remove app: unable to get git repo key: %s", err.Error())
	}

	repo, err := a.repoManager.Get(context.Background(), key, sourceRepo.Spec.URL, "test")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "remove app: unable to get git repo: %s", err.Error())
	}

	err = a.remover.Remove(repo, key, msg.Name, types.FluxNamespace, msg.RepoName)
	if err == types.ErrNotFound {
		return &pb.RemoveAppResponse{
			Success: true,
		}, nil
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to remove app: %s", err.Error())
	}

	return &pb.RemoveAppResponse{
		Success: true,
	}, nil
}
