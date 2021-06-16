package main

import (
	_ "oneTiny/config"
	"oneTiny/start"
)

func main() {
	start.Verify()
	start.Start()
}
