package main

import (
	"log"
	"os"

	"github.com/TCP404/OneTiny-cli/cmd"
	"github.com/TCP404/OneTiny-cli/config"
	"github.com/TCP404/OneTiny-cli/core"
	"github.com/TCP404/OneTiny-cli/external"
	"github.com/fatih/color"
)

func main() {
	var err error
	defer func() {
		if err != nil {
			log.Println(color.RedString("%v", err))
		}
	}()

	if err = config.LoadConfig(); err != nil {
		return
	}
	if err = cmd.CLI().Run(os.Args); err != nil {
		return
	}
	coreClient := external.Core
	if err = coreClient.NewVerifyChain().
		AddToHead(coreClient.NewPortHandler(config.Port)).
		AddToHead(coreClient.NewPathHandler(config.RootPath)).
		Iterator(); err != nil {
		return
	}

	core.RunCore()
}
