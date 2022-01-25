package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
	"github.com/gempir/go-twitch-irc/v3"
)

var (
	DToken     string
	TToken     string
	TName      string
	consoleLog *log.Logger
	database   *bolt.DB
	tclient    *twitch.Client
	channels   map[string]Channel
	library    map[string]discordgo.MessageEmbed
	guilds     map[string]bool
)

func init() {
	flag.StringVar(&DToken, "d", "", "Discord Token")
	flag.StringVar(&TToken, "t", "", "Twitch Token")
	flag.StringVar(&TName, "n", "", "Twitch Username")
	flag.Parse()
	LoadConfigFile()
	consoleLog = log.Default()
	database, _ = bolt.Open("data.db", 0600, nil)
	CreateBuckets()
	LoadChannels()
	LoadLibrary()
	LoadGuilds()
}

func main() {
	defer database.Close()

	//Setting Discord Up
	dg, err := discordgo.New("Bot " + DToken)
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
	tclient = twitch.NewClient(TName, TToken)
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
