package types

import "errors"

const (
	baseDir       = ".weave-gitops"
	FluxNamespace = "wego-system"
)

var (
	ErrNotFound = errors.New("entity not found")
)
