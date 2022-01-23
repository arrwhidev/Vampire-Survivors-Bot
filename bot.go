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
	channels   map[string]Channel
	library    map[string]discordgo.MessageEmbed
	guilds     map[string]bool
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
	consoleLog = log.Default()
	database, _ = bolt.Open("data.db", 0600, nil)
	CreateBuckets()
	LoadChannels()
	LoadLibrary()
	LoadGuilds()
}

func main() {
	defer database.Close()

	dg, err := discordgo.New("Bot " + Token)
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

	consoleLog.Println("[SETUP] Now running.  Press CTRL-C to exit.")
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
	//Admin commands
	if admin, _ := IsAdmin(s, m); admin {
		if strings.HasPrefix(m.Content, "!setvamp") {
			ch, _ := CreateChan(m.ChannelID, "")
			channels[m.ChannelID] = ch
			if _, ok := guilds[m.GuildID]; !ok {
				guilds[m.GuildID] = true
				CreateGuild(m.GuildID)
			}
			return
		}
	}
	//Regular commands
	if ch, ok := channels[m.ChannelID]; ok {
		if strings.HasPrefix(m.Content, ch.Prefix) {
			args := m.Content[len(ch.Prefix)+1:]
			if embd, ok := library[strings.ToLower(args)]; ok {
				err := SendEmbed(s, m.ChannelID, embd)
				if err != nil {
					consoleLog.Printf("[CMD] Command %s Failed! %v", args, err)
				} else {
					consoleLog.Printf("[CMD] Command %s Successful!", args)
				}
				return
			}
		}
	}
}

//Handles addition to new Guilds
func guildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	var firstAvailable *discordgo.Channel
	if _, ok := guilds[g.ID]; ok {
		return
	}
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
