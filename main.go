package main

import (
	_ "oneTiny/config"
	"oneTiny/start"
)

func main() {
	start.RunCLI()
	start.Verify()
	start.RunCore()
}
