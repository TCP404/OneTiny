package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	c.Redirect(http.StatusPermanentRedirect, "/file/")
}

func errorHandle(c *gin.Context, msg string) {
	c.String(http.StatusInternalServerError, msg)
}

func NotFound(c *gin.Context) {
	c.String(http.StatusNotFound, "404 Page Not Found", nil)
}
