package server

import (
	"context"
	"fmt"

	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"k8s.io/apimachinery/pkg/types"
)

type server struct {
	pb.UnimplementedApplicationsServer

	apps app.AppService
}

func NewApplicationsServer(apps app.AppService) pb.ApplicationsServer {
	return &server{
		apps: apps,
	}
}

func (s *server) ListApplications(ctx context.Context, msg *pb.ListApplicationsRequest) (*pb.ListApplicationsResponse, error) {
	apps, err := s.apps.List(ctx, msg.GetNamespace())
	if err != nil {
		return nil, err
	}

	if apps == nil {
		return &pb.ListApplicationsResponse{
			Applications: []*pb.Application{},
		}, nil
	}

	list := []*pb.Application{}
	for _, a := range apps {
		list = append(list, &pb.Application{Name: a.Name})
	}
	return &pb.ListApplicationsResponse{
		Applications: list,
	}, nil
}

func (s *server) GetApplication(ctx context.Context, msg *pb.GetApplicationRequest) (*pb.GetApplicationResponse, error) {
	app, err := s.apps.Get(ctx, types.NamespacedName{Name: msg.Name, Namespace: msg.Namespace})

	if err != nil {
		return nil, fmt.Errorf("could not get application: %s", err)
	}

	return &pb.GetApplicationResponse{Application: &pb.Application{
		Name: app.Name,
		Url:  app.Spec.URL,
		Path: app.Spec.Path,
	}}, nil
}

func (s *server) AddApplication(ctx context.Context, msg *pb.AddApplicationRequest) (*pb.AddApplicationResponse, error) {
	addParams := app.AddParams{
		Name:           msg.Name,
		Namespace:      msg.Namespace,
		Url:            msg.Url,
		Path:           msg.Path,
		Branch:         msg.Branch,
		DeploymentType: app.DeploymentType(msg.DeploymentType.String()),
		Chart:          msg.Chart,
		SourceType:     app.SourceType(msg.SourceType.String()),
		AppConfigUrl:   msg.AppConfigUrl,
		DryRun:         msg.DryRun,
		AutoMerge:      msg.AutoMerge,
	}

	if err := s.apps.Add(ctx, addParams); err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("could not add app: %w", err)
	}

	return &pb.AddApplicationResponse{}, nil
}
