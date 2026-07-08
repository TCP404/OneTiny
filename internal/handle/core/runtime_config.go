package core

import (
	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/state"
)

func currentSnapshot(c *gin.Context) state.ConfigSnapshot {
	if c != nil {
		if value, ok := c.Get(state.ContextKey); ok {
			if snapshot, ok := value.(state.ConfigSnapshot); ok {
				return snapshot
			}
		}
	}

	cfg := state.Current()
	if cfg != nil {
		return cfg.Snapshot()
	}

	return state.SnapshotFromCurrentConfig()
}
