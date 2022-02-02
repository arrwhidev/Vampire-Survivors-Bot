package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/gempir/go-twitch-irc/v3"
)

var (
	consoleLog *log.Logger
	database   *bolt.DB
	tclient    *twitch.Client
	channels   map[string]Channel
	library    map[string]discordgo.MessageEmbed
	guilds     map[string]bool
	aliases    map[string]string
)

func init() {
	consoleLog = log.Default()
	err := godotenv.Load()
	if err != nil {
		consoleLog.Fatal("Could not load environment")
	}
	database, _ = bolt.Open("data.db", 0600, nil)
	CreateBuckets()
	LoadChannels()
	LoadLibrary()
	LoadAliases()
	LoadGuilds()
}

func main() {
	defer database.Close()

	//Setting Discord Up
	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		consoleLog.Println("[SETUP] Error creating Discord session,", err)
		return
	}
	dg.AddHandler(messageCreate)
	dg.AddHandler(guildCreate)
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessages
	err = dg.Open()
	if err != nil {
		consoleLog.Println("[SETUP] Error opening connection,", err)
		return
	}

	//Setting Twitch client
	tclient = twitch.NewClient(os.Getenv("TWITCH_NAME"), os.Getenv("TWITCH_TOKEN"))
	tclient.OnPrivateMessage(twitchMessage)

	JoinInitialChans()

	go tclient.Connect()

	consoleLog.Println("[SETUP] Now running.  Press CTRL-C to exit.")

	//Gracefully close from console
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	consoleLog.Println("[SETUP] Shutting down...")
	dg.Close()
	tclient.Disconnect()
}
