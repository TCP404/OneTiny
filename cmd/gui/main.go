package main

import (
	"log"

	"github.com/tcp404/OneTiny/internal/conf"
	"github.com/tcp404/OneTiny/internal/gui"
	"github.com/tcp404/OneTiny/internal/gui/webassets"
)

func main() {
	if err := conf.LoadConfig(); err != nil {
		log.Fatal(err)
	}
	if err := gui.Run(webassets.Assets); err != nil {
		log.Fatal(err)
	}
}
