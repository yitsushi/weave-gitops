package router

import (
	"github.com/gin-gonic/gin"
)

func NewRestEngine() *gin.Engine {
	engine := gin.Default()

	engine.POST("/gitops/source/event", eventSourceHandler)

	engine.GET("/repository/:name", getRepoByName)

	engine.POST("/repository/:name/app", createApp)

	engine.GET("/repository/:name/artifact", getRepoArtifact)

	return engine
}
