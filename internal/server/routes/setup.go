package routes

import (
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/resource"
)

func Setup(r *gin.Engine) error {

	t, err := template.ParseFS(resource.FS, "template/*.tpl")
	if err != nil {
		return err
	}
	r.SetHTMLTemplate(t)

	loadIndexRoute(r)
	loadCoreRoute(r)
	loadLoginRoute(r)
	load404Route(r)
	loadICORoute(r)
	loadLogoRoute(r)
	return nil
}
