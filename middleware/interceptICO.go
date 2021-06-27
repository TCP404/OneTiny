package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func InterceptICO(c *gin.Context) {
	if strings.HasSuffix(c.Param("filename"), ".ico") { // 拦截浏览器默认请求 favicon.ico 的行为
		c.Status(http.StatusOK)
		c.Abort()
	}
}
