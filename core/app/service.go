package app

import (
	"github.com/weaveworks/weave-gitops/core/source"
	"k8s.io/apimachinery/pkg/util/uuid"
)

type Service interface {
	Create(name, namespace, description string) App
	Get(name string) App
}

func NewService(sourceSvc source.Service) Service {
	return &defaultService{
		sourceSvc: sourceSvc,
	}
}

type defaultService struct {
	sourceSvc source.Service
}

func (d defaultService) Create(name, namespace, description string) App {
	app := App{
		Id:          string(uuid.NewUUID()),
		Name:        name,
		Namespace:   namespace,
		Description: description,
	}

	return app
}

func (d defaultService) Get(name string) App {
	panic("implement me")
}
