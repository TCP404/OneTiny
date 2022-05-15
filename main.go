package main

import (
	"log"
	"oneTiny/cmd"
	"oneTiny/common/config"
	"oneTiny/common/verify"
	"oneTiny/core"
	"os"

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
	if err = cmd.RunCLI().Run(os.Args); err != nil {
		return
	}
	if err = verify.Verify(); err != nil {
		return
	}

	core.RunCore()
}
