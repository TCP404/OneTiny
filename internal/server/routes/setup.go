package routes

import (
	"html/template"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/resource"
)

func Setup(r *gin.Engine) *gin.Engine {

	t, err := template.ParseFS(resource.FS, "template/*.tpl")
	if err != nil {
		log.Fatal(err.Error())
	}
	r.SetHTMLTemplate(t)

	loadIndexRoute(r)
	loadCoreRoute(r)
	loadLoginRoute(r)
	load404Route(r)
	loadICORoute(r)
	loadLogoRoute(r)
	return r
}
