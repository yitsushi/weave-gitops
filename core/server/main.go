package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/weaveworks/weave-gitops/core/gitops"
	"github.com/weaveworks/weave-gitops/core/source"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"golang.org/x/sync/errgroup"
)

var (
	g errgroup.Group
)

func main() {
	engine := gin.Default()

	engine.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	engine.GET("/repository/:name", func(c *gin.Context) {
		_, client, err := kube.NewKubeHTTPClient()
		if err != nil {
			c.Error(err)
		}
		repo := source.NewService(client, source.GitopsRuntimeExclusionList)
		k, err := repo.Get(context.Background(), c.Param("name"), gitops.FluxNamespace)
		if err != nil {
			c.Error(err)
		}

		c.JSON(http.StatusOK, k)
	})

	engine.GET("/repository/:name/artifact", func(c *gin.Context) {
		_, client, err := kube.NewKubeHTTPClient()
		if err != nil {
			_ = c.Error(err)
		}
		repo := source.NewService(client, source.GitopsRuntimeExclusionList)
		k, err := repo.GetArtifact(context.Background(), c.Param("name"), gitops.FluxNamespace)
		if err != nil {
			_ = c.Error(err)
		}

		c.JSON(http.StatusOK, k)
	})

	g.Go(func() error {
		return engine.Run()
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("unable to start gin routers: %s", err.Error())
	}
}
