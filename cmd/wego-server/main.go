package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
)

func init() {
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		// Only log the debug severity or above.
		log.SetLevel(log.DebugLevel)
	} else if os.Getenv("LOG_LEVEL") == "WARN" {
		// Only log the warning severity or above.
		log.SetLevel(log.WarnLevel)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	command := NewAPIServerCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}

func NewAPIServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "wego-server",
		Long: `The wego-server handles HTTP requests for Weave GitOps Applications`,

		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return StartServer()
		},
	}
	return cmd
}

var addr = "0.0.0.0:8000"

func StartServer() error {
	log.Infof("wego api server started. listen address: %s", addr)
	return RunInProcessGateway(context.Background(), addr)
}

// RunInProcessGateway starts the invoke in process http gateway.
func RunInProcessGateway(ctx context.Context, addr string, opts ...runtime.ServeMuxOption) error {
	mux := runtime.NewServeMux(opts...)

	kubeClient, err := kube.NewKubeHTTPClient()
	if err != nil {
		log.Fatalf("could not create http client: %s", err)
	}

	gitClient := git.New(nil)
	fluxClient := flux.New(&runner.CLIRunner{})
	gitProviderClient := gitproviders.New()
	logger := logger.New(os.Stdout)
	appSvc := app.New(logger, gitClient, fluxClient, kubeClient, gitProviderClient)

	if err := pb.RegisterApplicationsHandlerServer(ctx, mux, server.NewApplicationsServer(appSvc)); err != nil {
		return fmt.Errorf("could not register application: %w", err)
	}
	s := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		log.Infof("Shutting down the http gateway server")
		if err := s.Shutdown(context.Background()); err != nil {
			log.Errorf("Failed to shutdown http gateway server: %v", err)
		}
	}()

	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Errorf("Failed to listen and serve: %v", err)
		return err
	}
	return nil
}
