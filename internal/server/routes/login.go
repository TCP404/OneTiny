package routes

import (
	"github.com/tcp404/OneTiny/internal/handle/secure"

	"github.com/gin-gonic/gin"
)

func loadLoginRoute(r *gin.Engine) {
	r.GET("/login", secure.LoginGet)
	r.POST("/login", secure.LoginPost)
}
