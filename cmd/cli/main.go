package main

import (
	"log"
	"os"

	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/kit/chain"
	"github.com/tcp404/OneTiny/internal/kit/verify"
	"github.com/tcp404/OneTiny/internal/server"
	"github.com/tcp404/OneTiny/internal/state"

	"github.com/fatih/color"
)

func main() {
	var err error
	defer func() {
		if err != nil {
			log.Println(color.RedString("%v", err))
		}
	}()

	cfg, err := conf.LoadConfig()
	if err != nil {
		return
	}
	state := state.NewRuntimeConfig(state.SnapshotFromConfig(cfg, state.NewProcessState()))

	if err = CLI(state).Run(os.Args); err != nil {
		return
	}

	snapshot := state.Snapshot()
	if err = conf.ValidateSecureConfigFor(snapshot.IsSecure); err != nil {
		return
	}

	if err = chain.NewHandleChain().
		AddToHead(verify.NewPortVerifier(snapshot.Port)).
		AddToHead(verify.NewPathVerifier(snapshot.RootPath)).
		Iterator(); err != nil {
		return
	}

	server.RunCore(state)
}
