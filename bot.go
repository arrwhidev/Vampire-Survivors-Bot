package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
)

//https://discord.com/api/oauth2/authorize?client_id=761955552091701258&permissions=52224&scope=bot

var (
	Token      string
	consoleLog *log.Logger
	database   *bolt.DB
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	consoleLog = log.Default()
	database, _ = bolt.Open("data.db", 0600, nil)
	defer database.Close()
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		consoleLog.Println("[BOT] Error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.AddHandler(guildCreate)

	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		consoleLog.Println("[BOT] Error opening connection,", err)
		return
	}

	consoleLog.Println("[BOT] Now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

//Handles messages
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if admin, _ := IsAdmin(s, m); admin {
	}
}

//Handles addition to new Guilds
func guildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	var firstAvailable *discordgo.Channel
	for _, channel := range g.Channels {
		if strings.Contains(channel.Name, "general") {
			firstAvailable = channel
			break
		}
		p, err := s.UserChannelPermissions(s.State.User.ID, channel.Name)
		if err != nil || p&discordgo.PermissionSendMessages == 0 {
			continue
		} else {
			firstAvailable = channel
		}
	}
	if firstAvailable != nil {
		s.ChannelMessageSend(firstAvailable.ID, "Yayaya")
	}
}
