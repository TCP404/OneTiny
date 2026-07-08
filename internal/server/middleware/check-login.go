package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/tcp404/OneTiny/internal/state"
)

func CheckLogin(c *gin.Context) {
	// 检查 session，
	// 有则放行
	// 无则跳转登录页面
	cfg := currentSnapshot()
	if !cfg.IsSecure {
		return
	}

	if session := sessions.Default(c); session.Get("login") == sessionVal() {
		c.Next()
		return
	} else {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}
}

func sessionVal() string {
	cfg := state.Current()
	if cfg == nil {
		return state.SnapshotFromCurrentConfig().SessionVal
	}

	if runtimeVal := cfg.Snapshot().SessionVal; runtimeVal != "" {
		return runtimeVal
	}
	return state.SnapshotFromCurrentConfig().SessionVal
}
