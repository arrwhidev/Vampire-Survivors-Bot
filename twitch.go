package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gempir/go-twitch-irc/v3"
)

//Handles incoming twitch messages
func twitchMessage(m twitch.PrivateMessage) {
	//Checking self channel for join requests
	if strings.EqualFold(m.Channel, os.Getenv("TWITCH_NAME")) {
		if m.Message == "!biteme" || m.Message == "!setvamp" {
			ch, _ := CreateChan(strings.ToLower(m.User.Name), "!")
			channels[strings.ToLower(m.User.Name)] = ch
			tclient.Join(strings.ToLower(m.User.Name))
			tclient.Say(m.Channel, fmt.Sprintf("Ouch! You've been bitten! I will now respond to commands in %v!", m.User.Name))
		}
		return
	}
	//Regular commands
	if ch, ok := channels[m.Channel]; ok {
		if strings.HasPrefix(m.Message, ch.Prefix) {
			args := m.Message[len(ch.Prefix):]
			if embd, ok := library[strings.ToLower(args)]; ok {
				tclient.Say(m.Channel, createResponse(embd))
				consoleLog.Printf("[CMD] Command %s Successful!", args)
				return
			}
			if alias, ok := aliases[strings.ToLower(args)]; ok {
				embd := library[alias]
				tclient.Say(m.Channel, createResponse(embd))
				consoleLog.Printf("[CMD] Command %s Successful!", args)
				return
			}
		}
		return
	}
}

//Converts library content to Twitch appropriate message
func createResponse(content discordgo.MessageEmbed) string {
	var fields string
	for _, embed_f := range content.Fields {
		fields = fields + fmt.Sprintf("%s: %s. ", embed_f.Name, embed_f.Value)
	}
	fields = strings.Replace(fields, "\n", " ", -1)
	result := fmt.Sprintf("%s: %s | %s", content.Title, content.Description, fields)
	return strings.Replace(result, "||", "", -1)
}
