package routes

import (
	"oneTiny/core/controller"

	"github.com/gin-gonic/gin"
)

func loadLoginRoute(r *gin.Engine) {
	r.GET("/login", controller.LoginGet)
	r.POST("/login", controller.LoginPost)
}
