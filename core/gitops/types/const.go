package types

import "errors"

const (
	BaseDir       = ".weave-gitops"
	FluxNamespace = "wego-system"
)

var (
	ErrNotFound = errors.New("entity not found")
)
