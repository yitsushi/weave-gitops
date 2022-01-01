package router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/weaveworks/weave-gitops/core/gitops"
	"github.com/weaveworks/weave-gitops/core/gitops/types"
	"github.com/weaveworks/weave-gitops/core/repository"
	"github.com/weaveworks/weave-gitops/core/source"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

type appRequest struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
}

func createApp(c *gin.Context) {
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

	repo, err := repoSvc.Get(context.Background(), c.Param("name"), types.FluxNamespace)
	if err != nil {
		c.Error(err)
	}

	gitClient, err := repoSvc.GitClient(context.Background(), types.FluxNamespace, repo)
	if err != nil {
		c.Error(err)
	}

	gitSvc := repository.NewGitWriter(gitClient, repo)
	appSvc := gitops.NewAppService(gitSvc)

	app, err := appSvc.Create(b.Name, b.Namespace, b.Description)
	if err != nil {
		c.Error(err)
	}

	c.JSON(http.StatusOK, app)
}
