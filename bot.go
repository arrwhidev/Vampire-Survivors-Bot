package main

import (
	"log"

	"github.com/joho/godotenv"
)

type Config struct {
	DToken string
	TToken string
	TName  string
	Owner  string
	Prefix string
}

type VampBot struct {
	Logger   *log.Logger
	Creds    Config
	Discord  *DiscordHandler
	Twitch   *TwitchHandler
	Library  *LibraryHandler
	Database *DatabaseHandler
}

func MakeVampBot(args Config) *VampBot {
	This := &VampBot{Creds: args}
	This.Logger = log.Default()
	This.Database = MakeDatabaseHandler(This, "data.db")
	This.Library = MakeLibraryHandler(This)
	This.Discord = MakeDiscordHandler(This)
	This.Twitch = MakeTwitchHandler(This)
	return This
}

func (bot *VampBot) Start() {
	bot.Discord.Start()
	bot.Twitch.Start()
	bot.Logger.Println("[SETUP] Now running.  Press CTRL-C to exit.")
}

func (bot *VampBot) Stop() {
	bot.Logger.Println("[SETUP] Shutting down...")
	bot.Discord.Stop()
	bot.Twitch.Stop()
}
func init() {
	godotenv.Load()
}
