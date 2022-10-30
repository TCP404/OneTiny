package routes

import (
	"html/template"
	"log"

	"github.com/TCP404/OneTiny-cli/resource"
	"github.com/gin-gonic/gin"
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
	return r
}
