package router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
	"github.com/weaveworks/weave-gitops/core/source"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

func getRepoByName(c *gin.Context) {
	_, client, err := kube.NewKubeHTTPClient()
	if err != nil {
		c.Error(err)
	}
	repo := source.NewService(client, source.GitopsRuntimeExclusionList)
	k, err := repo.Get(context.Background(), c.Param("name"), types.FluxNamespace)
	if err != nil {
		c.Error(err)
	}

	c.JSON(http.StatusOK, k)
}

func getRepoArtifact(c *gin.Context) {
	_, client, err := kube.NewKubeHTTPClient()
	if err != nil {
		_ = c.Error(err)
	}
	repo := source.NewService(client, source.GitopsRuntimeExclusionList)
	k, err := repo.GetArtifact(context.Background(), c.Param("name"), types.FluxNamespace)
	if err != nil {
		_ = c.Error(err)
	}

	c.JSON(http.StatusOK, k)
}
