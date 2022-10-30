package routes

import (
	"github.com/TCP404/OneTiny-cli/internal/handle/secure"

	"github.com/gin-gonic/gin"
)

func loadLoginRoute(r *gin.Engine) {
	r.GET("/login", secure.LoginGet)
	r.POST("/login", secure.LoginPost)
}
