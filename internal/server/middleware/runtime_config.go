package middleware

import (
	"github.com/tcp404/OneTiny/internal/state"
)

func currentSnapshot() state.ConfigSnapshot {
	cfg := state.Current()
	if cfg != nil {
		return cfg.Snapshot()
	}

	return state.SnapshotFromCurrentConfig()
}
