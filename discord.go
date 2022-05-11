package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

//https://discord.com/api/oauth2/authorize?client_id=761955552091701258&permissions=52224&scope=bot

type DiscordHandler struct {
	Bot     *VampBot
	Session *discordgo.Session
}

func MakeDiscordHandler(bot *VampBot) *DiscordHandler {
	dg, err := discordgo.New("Bot " + bot.Creds.DToken)
	if err != nil {
		bot.Logger.Fatal(err)
	}
	h := &DiscordHandler{Bot: bot, Session: dg}
	dg.AddHandler(h.messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessages
	return h
}

func (h *DiscordHandler) Start() {
	err := h.Session.Open()
	if err != nil {
		h.Bot.Logger.Fatal(err)
	}
	h.Bot.Logger.Println("[SETUP] Connected to Discord")
}

func (h *DiscordHandler) Stop() {
	h.Session.Close()
}

//Handles messages
func (h *DiscordHandler) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	//Admin commands
	if admin, _ := IsAdmin(s, m); admin {
		if strings.HasPrefix(m.Content, "!setvamp") {
			ch, _ := h.Bot.Database.CreateChan(m.ChannelID, "!")
			h.Bot.Database.Chans[m.ChannelID] = ch
			if _, ok := h.Bot.Database.Guilds[m.GuildID]; !ok {
				h.Bot.Database.Guilds[m.GuildID] = true
				h.Bot.Database.CreateGuild(m.GuildID)
			}
			s.ChannelMessageSend(m.ChannelID, "Hello! I will now respond to commands here! Try `!garlic`")
			return
		}
	}
	//Regular commands
	if ch, ok := h.Bot.Database.Chans[m.ChannelID]; ok {
		if strings.HasPrefix(m.Content, ch.Prefix) {
			args := m.Content[len(ch.Prefix):]
			if embd, ok := h.Bot.Library.GetItem(strings.ToLower(args), false); ok {
				err := SendEmbed(s, m.ChannelID, embd)
				if err != nil {
					h.Bot.Logger.Printf("[CMD] Command %s Failed! %v", args, err)
				} else {
					h.Bot.Logger.Printf("[CMD] Command %s Successful!", args)
				}
				return
			}
		}
	}
}

func SendEmbed(s *discordgo.Session, c string, m discordgo.MessageEmbed) error {
	send := &discordgo.MessageSend{
		Embed: &m,
		TTS:   false,
	}
	_, err := s.ChannelMessageSendComplex(c, send)
	return err
}

//Checks wether message author has administrator permissions
func IsAdmin(s *discordgo.Session, m *discordgo.MessageCreate) (bool, error) {
	perm, err := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
	if err != nil {
		return false, err
	}
	return perm&discordgo.PermissionAdministrator != 0, nil
}
