package main

import (
	"log"
	"os"

	"github.com/TCP404/OneTiny-cli/cmd"
	"github.com/TCP404/OneTiny-cli/internal/conf"
	"github.com/TCP404/OneTiny-cli/internal/kit/verify"
	"github.com/TCP404/OneTiny-cli/internal/server"
	"github.com/TCP404/OneTiny-cli/pkg/container"

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

	if err = cmd.CLI().Run(os.Args); err != nil {
		return
	}

	if err = container.NewHandleChain().
		AddToHead(verify.NewPortVerifier(conf.Config.Port)).
		AddToHead(verify.NewPathVerifier(conf.Config.RootPath)).
		Iterator(); err != nil {
		return
	}

	server.RunCore()
}
