package routes

import (
	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) *gin.Engine {
	loadIndexRoute(r)
	loadCoreRoute(r)
	loadLoginRoute(r)
	load404Route(r)
	loadICORoute(r)
	return r
}
