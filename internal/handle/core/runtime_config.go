package core

import (
	"github.com/TCP404/OneTiny-cli/internal/conf"
	"github.com/TCP404/OneTiny-cli/internal/runtimeconf"
	"github.com/gin-gonic/gin"
)

func currentSnapshot(c *gin.Context) runtimeconf.ConfigSnapshot {
	if c != nil {
		if value, ok := c.Get(runtimeconf.ContextKey); ok {
			if snapshot, ok := value.(runtimeconf.ConfigSnapshot); ok {
				return snapshot
			}
		}
	}

	cfg := runtimeconf.Current()
	if cfg != nil {
		return cfg.Snapshot()
	}

	return runtimeconf.ConfigSnapshot{
		RootPath:      conf.Config.RootPath,
		Port:          conf.Config.Port,
		MaxLevel:      conf.Config.MaxLevel,
		IsAllowUpload: conf.Config.IsAllowUpload,
		IsSecure:      conf.Config.IsSecure,
		IP:            conf.Config.IP,
	}
}
