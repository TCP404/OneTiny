package handle

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandle(c *gin.Context, msg string) {
	c.String(http.StatusInternalServerError, msg)
}
