package main

import (
	_ "oneTiny/config"
	"oneTiny/start"
)

func main() {
	start.Run()
	start.Verify()
	start.StartGin()
}
