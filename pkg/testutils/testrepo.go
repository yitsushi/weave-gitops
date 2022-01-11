package testutils

import (
	"time"

	"github.com/fluxcd/pkg/gittestserver"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func SetupGitServer(auth *http.BasicAuth) (*gittestserver.GitServer, error) {
	gitServer, err := gittestserver.NewTempGitServer()

	if err != nil {
		return nil, err
	}

	gitServer.Auth(auth.Username, auth.Password)

	errc := make(chan error)

	go func() {
		errc <- gitServer.StartHTTP()
	}()

	select {
	case err := <-errc:
		if err != nil {
			return nil, err
		}
	case <-time.After(time.Second):
		break
	}

	return gitServer, nil
}
