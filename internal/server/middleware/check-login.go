package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func CheckLogin(c *gin.Context) {
	// 检查 session，
	// 有则放行
	// 无则跳转登录页面
	cfg := currentSnapshot(c)
	if !cfg.IsSecure {
		return
	}

	if session := sessions.Default(c); session.Get("login") == cfg.SessionVal {
		c.Next()
		return
	} else {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		c.Abort()
		return
	}
}
