package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	bot := MakeVampBot(Config{})
	bot.Start()

	//Gracefully close from console
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	bot.Stop()
}
