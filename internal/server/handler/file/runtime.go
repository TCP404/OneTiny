package file

import (
	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/runtime"
)

func currentSnapshot(c *gin.Context) runtime.Snapshot {
	if c == nil {
		return runtime.Snapshot{}
	}
	value, ok := c.Get(runtime.ContextKey)
	if !ok {
		return runtime.Snapshot{}
	}
	snapshot, ok := value.(runtime.Snapshot)
	if !ok {
		return runtime.Snapshot{}
	}
	return snapshot
}
