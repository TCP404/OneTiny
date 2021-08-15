package middleware

import (
	"net/http"
	"oneTiny/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func CheckLogin(c *gin.Context) {
	// 检查 session，
	// 有则放行
	// 无则跳转登录页面
	if !config.IsSecure {
		return
	}

	if session := sessions.Default(c); session.Get("login") == config.SessionVal {
		c.Next()
		return
	} else {
		c.Redirect(http.StatusMovedPermanently, "/login")
		return
	}
}
