package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

//https://discord.com/api/oauth2/authorize?client_id=761955552091701258&permissions=52224&scope=bot

//Handles messages
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	//Admin commands
	if admin, _ := IsAdmin(s, m); admin {
		if strings.HasPrefix(m.Content, "!setvamp") {
			ch, _ := CreateChan(m.ChannelID, "!")
			channels[m.ChannelID] = ch
			if _, ok := guilds[m.GuildID]; !ok {
				guilds[m.GuildID] = true
				CreateGuild(m.GuildID)
			}
			s.ChannelMessageSend(m.ChannelID, "Hello! I will now respond to commands here! Try `!garlic`")
			return
		}
	}
	//Regular commands
	if ch, ok := channels[m.ChannelID]; ok {
		if strings.HasPrefix(m.Content, ch.Prefix) {
			args := m.Content[len(ch.Prefix):]
			if embd, ok := library[strings.ToLower(args)]; ok {
				err := SendEmbed(s, m.ChannelID, embd)
				if err != nil {
					consoleLog.Printf("[CMD] Command %s Failed! %v", args, err)
				} else {
					consoleLog.Printf("[CMD] Command %s Successful!", args)
				}
				return
			} 
			if alias, ok := aliases[strings.ToLower(args)]; ok {
				embd, _ := library[alias]
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
		s.ChannelMessageSend(firstAvailable.ID, "Hello! I am Vampire Turtle! A community library bot for Vampire Survivors!\nUse `!setvamp` in the channel you want me to respond to commands in!")
	}
}
