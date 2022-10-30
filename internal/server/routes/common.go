package routes

import (
	"net/http"

	"github.com/TCP404/OneTiny-cli/resource"
	"github.com/gin-gonic/gin"
)

func load404Route(app *gin.Engine) {
	app.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "404 Page Not Found", nil)
	})
}

func loadIndexRoute(app *gin.Engine) {
	app.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/file/")
	})
}

func loadICORoute(app *gin.Engine) {
	app.GET("favicon.ico", func(c *gin.Context) {
		file, _ := resource.FS.ReadFile("logo/favicon.ico")
		c.Data(http.StatusOK, "image/x-icon", file)
	})
}
