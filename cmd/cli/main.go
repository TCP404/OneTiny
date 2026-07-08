package main

import (
	"log"
	"os"

	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/kit/chain"
	"github.com/tcp404/OneTiny/internal/kit/verify"
	"github.com/tcp404/OneTiny/internal/server"

	"github.com/fatih/color"
)

func main() {
	var err error
	defer func() {
		if err != nil {
			log.Println(color.RedString("%v", err))
		}
	}()

	if err = conf.LoadConfig(); err != nil {
		return
	}

	if err = CLI().Run(os.Args); err != nil {
		return
	}

	if err = conf.ValidateSecureConfigFor(conf.Config.IsSecure); err != nil {
		return
	}

	if err = chain.NewHandleChain().
		AddToHead(verify.NewPortVerifier(conf.Config.Port)).
		AddToHead(verify.NewPathVerifier(conf.Config.RootPath)).
		Iterator(); err != nil {
		return
	}

	server.RunCore()
}
