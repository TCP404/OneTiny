package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) *gin.Engine {
	r.Use(Logger(), gin.Recovery())
	r.Use(enableCookieSession())
	return r
}

func enableCookieSession() gin.HandlerFunc {
	s := cookie.NewStore([]byte("secret"))
	return sessions.Sessions("SESSIONID", s)
}
