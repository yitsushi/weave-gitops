package router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/weaveworks/weave-gitops/core/gitops/app"

	"github.com/weaveworks/weave-gitops/core/gitops/types"
	"github.com/weaveworks/weave-gitops/core/source"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

type appRequest struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
}

func createApp(c *gin.Context) {
	//var b appRequest
	//
	//err := c.BindJSON(&b)
	//if err != nil {
	//	c.String(http.StatusBadRequest, "%s", err)
	//}
	//
	//_, client, err := kube.NewKubeHTTPClient()
	//if err != nil {
	//	_ = c.Error(err)
	//}
	//
	//repoSvc := source.NewService(client, source.GitopsRuntimeExclusionList)
	//
	//repo, err := repoSvc.Get(context.Background(), c.Param("name"), types.FluxNamespace)
	//if err != nil {
	//	_ = c.Error(err)
	//}
	//
	//gitClient, err := repoSvc.GitClient(context.Background(), types.FluxNamespace, repo)
	//if err != nil {
	//	_ = c.Error(err)
	//}
	//
	//gitSvc := repository.NewGitWriter(gitClient, repo)
	//appSvc := app.NewCreator(gitSvc)
	//
	//app, err := appSvc.Create(b.Name, b.Namespace, b.Description, "delta")
	//if err != nil {
	//	_ = c.Error(err)
	//}
	//
	//c.JSON(http.StatusOK, app)
}

func deleteApp(c *gin.Context) {
	//var b appRequest
	//
	//err := c.BindJSON(&b)
	//if err != nil {
	//	c.String(http.StatusBadRequest, "%s", err)
	//}
	//
	//_, client, err := kube.NewKubeHTTPClient()
	//if err != nil {
	//	_ = c.Error(err)
	//}
	//
	//repoSvc := source.NewService(client, source.GitopsRuntimeExclusionList)
	//
	//repo, err := repoSvc.Get(context.Background(), c.Param("name"), types.FluxNamespace)
	//if err != nil {
	//	_ = c.Error(err)
	//}
	//
	//gitClient, err := repoSvc.GitClient(context.Background(), types.FluxNamespace, repo)
	//if err != nil {
	//	_ = c.Error(err)
	//}
	//
	//gitSvc := repository.NewGitWriter(gitClient, repo)
	//appFetcher := app.NewFetcher(repoSvc)
	//
	//appObj, err := appFetcher.Get(context.Background(), c.Param("appName"), c.Param("name"), types.FluxNamespace)
	//if err != nil {
	//	_ = c.Error(err)
	//}
	//
	//appRemover := app.NewRemover(gitSvc)
	//
	//err = appRemover.Remove(appObj, "delta")
	//if err != nil {
	//	_ = c.Error(err)
	//}
	//
	//c.String(http.StatusNoContent, "")
}

func listApps(c *gin.Context) {
	_, client, err := kube.NewKubeHTTPClient()
	if err != nil {
		_ = c.Error(err)
	}

	sourceSvc := source.NewService(client, source.GitopsRuntimeExclusionList)
	appFetcher := app.NewFetcher(sourceSvc)

	apps, err := appFetcher.List(context.Background(), c.Param("name"), types.FluxNamespace)
	if err != nil {
		_ = c.Error(err)
	}

	if len(apps) == 0 {
		c.String(http.StatusOK, "[]")
	} else {
		c.JSON(http.StatusOK, apps)
	}
}

func getApp(c *gin.Context) {
	_, client, err := kube.NewKubeHTTPClient()
	if err != nil {
		_ = c.Error(err)
	}

	sourceSvc := source.NewService(client, source.GitopsRuntimeExclusionList)
	appFetcher := app.NewFetcher(sourceSvc)

	app, err := appFetcher.Get(context.Background(), c.Param("appName"), c.Param("name"), types.FluxNamespace)
	if err != nil {
		_ = c.Error(err)
	}

	c.JSON(http.StatusOK, app)
}
