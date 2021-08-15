package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// InterceptICO 拦截浏览器默认请求 favicon.ico 的行为
func InterceptICO(c *gin.Context) {
	if strings.HasSuffix(c.Param("filename"), ".ico") {
		c.Status(http.StatusOK)
		c.Abort()
	}
}
