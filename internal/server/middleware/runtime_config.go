package middleware

import (
	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/runtimeconf"
)

func currentSnapshot() runtimeconf.ConfigSnapshot {
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
