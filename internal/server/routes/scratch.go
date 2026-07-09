package routes

import (
	"github.com/tcp404/OneTiny/internal/scratch"
	scratchhandler "github.com/tcp404/OneTiny/internal/server/handler/scratch"
	"github.com/tcp404/OneTiny/internal/server/middleware"
	"github.com/tcp404/OneTiny/internal/server/routepath"

	"github.com/gin-gonic/gin"
)

func loadScratchRoute(r *gin.Engine, store *scratch.Store) {
	group := r.Group(routepath.ScratchGroupPrefix, middleware.CheckLogin)
	scratchhandler.Register(group, store)
}
