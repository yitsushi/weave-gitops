package gitops

import "errors"

const (
	FluxNamespace = "wego-system"
)

var (
	ErrNotFound = errors.New("entity not found")
)
