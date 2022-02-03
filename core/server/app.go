package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/weaveworks/weave-gitops/api/v1alpha2"
	stypes "github.com/weaveworks/weave-gitops/core/server/types"
	"github.com/weaveworks/weave-gitops/core/services/deploykey"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Create the scheme once and re-use it on every call.
// This shouldn't need to change between requests(?)
var scheme = kube.CreateScheme()

type appServer struct {
	pb.UnimplementedAppsServer

	k8s  placeholderClientGetter
	http *http.Client
}

// This struct is only here to avoid a circular import with the `server` package.
// This is meant to match the ClientGetter interface.
// Since we are in a prototyping phase, it didn't make sense to move and import that code just yet.
type placeholderClientGetter struct {
	cfg *rest.Config
}

func (p placeholderClientGetter) Client(ctx context.Context) (client.Client, error) {
	return client.New(p.cfg, client.Options{
		Scheme: scheme,
	})
}

func NewAppServer(cfg *rest.Config) pb.AppsServer {
	return &appServer{
		k8s:  placeholderClientGetter{cfg: cfg},
		http: http.DefaultClient,
	}
}

func (as *appServer) AddApp(ctx context.Context, msg *pb.AddAppRequest) (*pb.AddAppResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	app := stypes.AppAddProtoToCustomResource(msg)

	err = k8s.Create(ctx, app)

	if k8serrors.IsUnauthorized(err) {
		return nil, status.Errorf(codes.PermissionDenied, err.Error())
	} else if k8serrors.IsNotFound(err) {
		return nil, status.Errorf(codes.NotFound, err.Error())
	} else if k8serrors.IsConflict(err) {

	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to create new app: %s", err.Error())
	}

	return &pb.AddAppResponse{
		App:     stypes.AppCustomResourceToProto(app),
		Success: true,
	}, nil
}

func (as *appServer) GetApp(ctx context.Context, msg *pb.GetAppRequest) (*pb.GetAppResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	obj := &v1alpha2.Application{}

	if err := k8s.Get(ctx, types.NamespacedName{Name: msg.AppName, Namespace: msg.Namespace}, obj); err != nil {
		return nil, status.Errorf(codes.Internal, "getting app: %s", err.Error())
	}

	return &pb.GetAppResponse{App: stypes.AppCustomResourceToProto(obj)}, nil
}

func (as *appServer) ListApps(ctx context.Context, msg *pb.ListAppRequest) (*pb.ListAppResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	list := &v1alpha2.ApplicationList{}

	err = k8s.List(ctx, list, client.InNamespace(msg.Namespace))
	if k8serrors.IsUnauthorized(err) {
		return nil, status.Errorf(codes.PermissionDenied, "")
	} else if k8serrors.IsNotFound(err) {
		return nil, status.Errorf(codes.NotFound, "")
	}

	var results []*pb.App
	for _, item := range list.Items {
		results = append(results, stypes.AppCustomResourceToProto(&item))
	}

	return &pb.ListAppResponse{
		Apps: results,
	}, nil
}

func (as *appServer) RemoveApp(_ context.Context, msg *pb.RemoveAppRequest) (*pb.RemoveAppResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "")
}

func (as *appServer) CreateDeployKey(ctx context.Context, msg *pb.CreateDeployKeyRequest) (*pb.CreateDeployKeyResponse, error) {
	token, err := middleware.ExtractProviderToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "extracting token: %s", err)
	}

	if msg.SecretName == "" {
		return nil, status.Errorf(codes.InvalidArgument, "secretName is a required field")
	}

	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	repoURL, err := gitproviders.NewRepoURL(msg.RepoUrl)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "creating repo url: %s", err)
	}

	name := types.NamespacedName{Name: msg.SecretName, Namespace: msg.Namespace}

	mgr := deploykey.NewManager(k8s, as.http.Transport)

	generatedName, err := mgr.Create(ctx, name, repoURL, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("creating deploy key: %s", err)
	}

	return &pb.CreateDeployKeyResponse{
		SecretName: string(generatedName),
	}, nil
}

func doClientError(err error) error {
	return status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
}
