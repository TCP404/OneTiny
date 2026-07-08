package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/resource"
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

func loadLogoRoute(app *gin.Engine) {
	app.GET("logo.png", func(c *gin.Context) {
		file, _ := resource.FS.ReadFile("logo/logo-white.png")
		c.Data(http.StatusOK, "image/png", file)
	})
}
