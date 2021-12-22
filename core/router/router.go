package router

import (
	"github.com/gin-gonic/gin"
)

func NewRestEngine() *gin.Engine {
	engine := gin.Default()
	engine.Use(gin.ErrorLogger())
	engine.Use(gin.Recovery())

	engine.POST("/gitops/source/event", eventSourceHandler)

	engine.GET("/repository/:name", getRepoByName)

	engine.GET("/repository/:name/app", listApps)
	engine.POST("/repository/:name/app", createApp)
	engine.GET("/repository/:name/app/:appName", getApp)
	engine.DELETE("/repository/:name/app/:appName", deleteApp)

	engine.GET("/repository/:name/artifact", getRepoArtifact)

	return engine
}
