package routes

import (
	"github.com/TCP404/OneTiny-cli/internal/controller"

	"github.com/gin-gonic/gin"
)

func loadLoginRoute(r *gin.Engine) {
	r.GET("/login", controller.LoginGet)
	r.POST("/login", controller.LoginPost)
}
