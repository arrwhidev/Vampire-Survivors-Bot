package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	bot := MakeVampBot(
		Config{
			DToken: os.Getenv("DISCORD_TOKEN"),
			TToken: os.Getenv("TWITCH_TOKEN"),
			TName:  os.Getenv("TWITCH_NAME"),
			Owner:  os.Getenv("OWNER_ID"),
			Prefix: os.Getenv("PREFIX"),
		})
	bot.Start()

	//Gracefully close from console
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	bot.Stop()
}
