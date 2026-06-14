package main

import (
	"log"

	"github.com/TCP404/OneTiny-cli/frontend"
	"github.com/TCP404/OneTiny-cli/internal/conf"
	"github.com/TCP404/OneTiny-cli/internal/gui"
)

func main() {
	if err := conf.LoadConfig(); err != nil {
		log.Fatal(err)
	}
	if err := gui.Run(frontend.Assets); err != nil {
		log.Fatal(err)
	}
}
