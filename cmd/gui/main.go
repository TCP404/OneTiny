package main

import (
	"log"

	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/gui"
	"github.com/tcp404/OneTiny/internal/gui/webassets"
	"github.com/tcp404/OneTiny/internal/state"
)

func main() {
	cfg, err := conf.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	state := state.NewRuntimeConfig(state.SnapshotFromConfig(cfg, state.NewProcessState()))
	if err := gui.Run(webassets.Assets, state); err != nil {
		log.Fatal(err)
	}
}
