package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func errorHandle(c *gin.Context, msg string) {
	c.String(http.StatusInternalServerError, msg)
}
