package server

import (
	"context"
	"fmt"

	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/core/gitops/app"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
	"github.com/weaveworks/weave-gitops/core/repository"
	"github.com/weaveworks/weave-gitops/core/source"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type kustServer struct {
	pb.UnimplementedAppKustomizationServer

	fetcher     app.Fetcher
	repoManager repository.Manager
	sourceSvc   source.Service
}

func protoToKustomization(kustomization *pb.AddKustomizationRequest) v1beta2.Kustomization {
	return v1beta2.Kustomization{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1beta2.KustomizationKind,
			APIVersion: v1beta2.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kustomization.Name,
			Namespace: kustomization.Namespace,
		},
		Spec: v1beta2.KustomizationSpec{
			Path:     kustomization.Path,
			Interval: intervalDuration(kustomization.Interval),
			SourceRef: v1beta2.CrossNamespaceSourceReference{
				Kind: kustomization.SourceRef.Kind.String(),
				Name: kustomization.Name,
			},
		},
		Status: v1beta2.KustomizationStatus{},
	}
}

func NewKustomizationServer(fetcher app.Fetcher, sourceSvc source.Service, repoManager repository.Manager) pb.AppKustomizationServer {
	return &kustServer{
		fetcher:     fetcher,
		repoManager: repoManager,
		sourceSvc:   sourceSvc,
	}
}

func (ks *kustServer) Add(ctx context.Context, msg *pb.AddKustomizationRequest) (*pb.AddKustomizationResponse, error) {
	repo, key, err := getRepo(ks.sourceSvc, ks.repoManager, msg.RepoName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "kustServer.Add: %w", err)
	}

	app, err := ks.fetcher.Get(ctx, msg.AppName, msg.RepoName, types.FluxNamespace)
	if err == types.ErrNotFound {
		return nil, status.Error(codes.NotFound, "resource does not exist")
	} else if err != nil {
		return nil, fmt.Errorf("kustServer.Add: %w")
	}

	app.AddFluxKustomization(protoToKustomization(msg))

	return &pb.AddKustomizationResponse{
		Success: true,
	}, nil
}

func (ks *kustServer) Remove(ctx context.Context, msg *pb.RemoveKustomizationRequest) (*pb.RemoveKustomizationResponse, error) {
	repo, key, err := getRepo(ks.sourceSvc, ks.repoManager, msg.RepoName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "kustServer.Remove: %w", err)
	}

	app, err := ks.fetcher.Get(ctx, msg.AppName, msg.RepoName, types.FluxNamespace)
	if err == types.ErrNotFound {
		return nil, status.Error(codes.NotFound, "resource does not exist")
	} else if err != nil {
		return nil, fmt.Errorf("kustServer.Add: %w")
	}

	return &pb.RemoveKustomizationResponse{
		Success: true,
	}, nil
}
