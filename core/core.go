package core

import (
	"oneTiny/core/middleware"
	"oneTiny/core/routes"

	"github.com/gin-gonic/gin"
)

func StartUpGin() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	middleware.Setup(r)
	routes.Setup(r)
	return r
}
