package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NotFound(c *gin.Context) {
	c.String(http.StatusNotFound, "404 Page Not Found", nil)
}

