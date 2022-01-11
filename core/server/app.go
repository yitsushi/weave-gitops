package server

import (
	"context"
	"errors"

	"github.com/weaveworks/weave-gitops/core/gitops/app"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
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

	fetcher   app.Fetcher
	sourceSvc source.Service
}

func NewAppServer(fetcher app.Fetcher, sourceSvc source.Service) pb.AppsServer {
	return &appServer{
		fetcher:   fetcher,
		sourceSvc: sourceSvc,
	}
}

func (a *appServer) AddApp(_ context.Context, msg *pb.AddAppRequest) (*pb.AddAppResponse, error) {
	//repo, err := a.sourceSvc.Get(context.Background(), msg.RepoName, types.FluxNamespace)
	//if err != nil {
	//	return nil, status.Errorf(codes.Internal, "unable to get config repo: %s", err.Error())
	//}
	//
	//gitClient, err := a.sourceSvc.GitClient(context.Background(), types.FluxNamespace, repo)
	//if err != nil {
	//	return nil, status.Errorf(codes.Internal, "unable to get git client: %s", err.Error())
	//}
	//
	//gitSvc := repository.NewGitWriter(gitClient, repo)
	//appSvc := app.NewCreator(gitSvc)
	//
	//app, err := appSvc.Create(msg.Name, types.FluxNamespace, msg.DisplayName, "delta")
	//if err != nil {
	//	return nil, status.Errorf(codes.Internal, "unable to create new app: %s", err.Error())
	//}

	//return &pb.AddAppResponse{App: newProtoApp(app)}, nil
	return nil, errors.New("not implemented")
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
	return nil, errors.New("not implemented")
}
