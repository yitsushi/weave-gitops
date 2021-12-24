package gitops

import (
	"github.com/weaveworks/weave-gitops/core/source"
	"k8s.io/apimachinery/pkg/util/uuid"
)

type AppService interface {
	Create(name, namespace, description string) App
	Get(name string) App
}

func NewAppService(sourceSvc source.Service) AppService {
	return &defaultAppService{
		sourceSvc: sourceSvc,
	}
}

type defaultAppService struct {
	sourceSvc source.Service
}

func (d defaultAppService) Create(name, namespace, description string) App {
	app := App{
		Id:          string(uuid.NewUUID()),
		Name:        name,
		Namespace:   namespace,
		Description: description,
	}

	return app
}

func (d defaultAppService) Get(name string) App {
	panic("implement me")
}
