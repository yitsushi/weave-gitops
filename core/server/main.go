package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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

	g.Go(func() error {
		return engine.Run()
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("unable to start gin routers: %s", err.Error())
	}
}
