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

type appRequest struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
}

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

	engine.POST("/repository/:name/app", func(c *gin.Context) {
		var b appRequest
		err := c.BindJSON(&b)
		if err != nil {
			c.String(http.StatusBadRequest, "%s", err)
		}

		_, client, err := kube.NewKubeHTTPClient()
		if err != nil {
			c.Error(err)
		}
		repoSvc := source.NewService(client, source.GitopsRuntimeExclusionList)
		repo, err := repoSvc.Get(context.Background(), c.Param("name"), gitops.FluxNamespace)
		if err != nil {
			c.Error(err)
		}

		gitClient, err := repoSvc.GitClient(context.Background(), repo)
		if err != nil {
			c.Error(err)
		}

		gitSvc := gitops.NewGitService(gitClient, repo)
		appSvc := gitops.NewAppService(gitSvc)

		app, err := appSvc.Create(b.Name, b.Namespace, b.Description)
		if err != nil {
			c.Error(err)
		}

		c.JSON(http.StatusOK, app)
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
