package routes

import (
	"github.com/tcp404/OneTiny/internal/server/handler/auth"

	"github.com/gin-gonic/gin"
)

func loadLoginRoute(r *gin.Engine) {
	r.GET("/login", auth.LoginGet)
	r.POST("/login", auth.LoginPost)
}
