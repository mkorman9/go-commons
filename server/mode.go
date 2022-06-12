package server

import (
	"github.com/gin-gonic/gin"
	"github.com/gookit/config/v2"
)

func init() {
	mode := config.String("server.mode")
	if mode == "" {
		mode = "release"
	}

	gin.SetMode(mode)
}
