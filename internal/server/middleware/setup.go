package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/accesslog"
	"github.com/tcp404/OneTiny/internal/runtime"
)

func Setup(r *gin.Engine, rt *runtime.Runtime, logger *accesslog.Logger) *gin.Engine {
	r.Use(Logger(rt), gin.Recovery())
	r.Use(RuntimeSnapshot(rt))
	r.Use(AccessLogger(logger))
	r.Use(enableCookieSession())
	r.Use(AccessLog(logger))
	return r
}

func enableCookieSession() gin.HandlerFunc {
	s := cookie.NewStore([]byte("secret"))
	return sessions.Sessions("SESSIONID", s)
}
