package source

import (
	"github.com/fluxcd/pkg/runtime/events"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
)

type EventHandler interface {
	InputEvent(event events.Event) error
}

func NewInMemoryEventHandler() EventHandler {
	return &gitRepositoryEventHandler{}
}

type gitRepositoryEventHandler struct {
}

func (gr gitRepositoryEventHandler) InputEvent(event events.Event) error {
	if event.InvolvedObject.Kind != sourcev1.GitRepositoryKind {
		return nil
	}

	return nil
}
