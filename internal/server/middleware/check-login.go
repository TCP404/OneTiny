package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/TCP404/OneTiny-cli/internal/conf"
)

func CheckLogin(c *gin.Context) {
	// 检查 session，
	// 有则放行
	// 无则跳转登录页面
	if !conf.Config.IsSecure {
		return
	}

	if session := sessions.Default(c); session.Get("login") == conf.Config.SessionVal {
		c.Next()
		return
	} else {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}
}
