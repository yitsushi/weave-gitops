package router

import (
	"fmt"

	"github.com/fluxcd/pkg/runtime/events"
	"github.com/gin-gonic/gin"
)

func eventSourceHandler(c *gin.Context) {
	var event events.Event
	err := c.BindJSON(&event)
	if err != nil {
		fmt.Printf("an error occurred: %w", err)
	} else {
		fmt.Printf("%+v", event)
	}

	fmt.Println()
}
