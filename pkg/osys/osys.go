package osys

import (
	"os"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Osys
type Osys interface {
	UserHomeDir() (string, error)
	Exit(code int)
}

type OsysClient struct{}

func New() *OsysClient {
	return &OsysClient{}
}

var _ Osys = &OsysClient{}

func (o *OsysClient) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (o *OsysClient) Exit(code int) {
	os.Exit(code)
}
