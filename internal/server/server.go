package server

import (
	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/accesslog"
	"github.com/tcp404/OneTiny/internal/runtime"
	"github.com/tcp404/OneTiny/internal/scratch"
	"github.com/tcp404/OneTiny/internal/server/middleware"
	"github.com/tcp404/OneTiny/internal/server/routes"
)

func setupEngine(r *gin.Engine, rt *runtime.Runtime, logger *accesslog.Logger, scratchStore *scratch.Store) error {
	middleware.Setup(r, rt, logger)
	return routes.Setup(r, scratchStore)
}
