package gitops

import "errors"

const (
	baseDir       = ".weave-gitops"
	FluxNamespace = "wego-system"
)

var (
	ErrNotFound = errors.New("entity not found")
)

type File struct {
	Path string
	Data []byte
}
