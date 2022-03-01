package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gempir/go-twitch-irc/v3"
)

type TwitchHandler struct {
	Bot     *VampBot
	Session *twitch.Client
}

func MakeTwitchHandler(bot *VampBot) *TwitchHandler {
	handler := &TwitchHandler{}
	handler.Session = twitch.NewClient(bot.Creds.TName, bot.Creds.TToken)
	handler.Session.OnPrivateMessage(handler.twitchMessage)
	handler.JoinInitialChans()
	return handler
}

func (handler *TwitchHandler) Start() {
	go handler.Session.Connect()
	handler.Bot.Logger.Println("[SETUP] Connecting to Twitch")
}

func (handler *TwitchHandler) Stop() {
	handler.Session.Disconnect()
}

var IsTChan = regexp.MustCompile(`[^0-9]`).FindString

//Handles incoming twitch messages
func (handler *TwitchHandler) twitchMessage(m twitch.PrivateMessage) {
	//Checking self channel for join requests
	if strings.EqualFold(m.Channel, handler.Bot.Creds.TName) {
		if m.Message == "!biteme" || m.Message == "!setvamp" {
			ch, _ := handler.Bot.Database.CreateChan(strings.ToLower(m.User.Name), "!")
			handler.Bot.Database.Chans[strings.ToLower(m.User.Name)] = ch
			handler.Session.Join(strings.ToLower(m.User.Name))
			handler.Session.Say(m.Channel, fmt.Sprintf("Ouch! You've been bitten! I will now respond to commands in %v!", m.User.Name))
		}
		return
	}
	//Regular commands
	if ch, ok := handler.Bot.Database.Chans[m.Channel]; ok {
		if strings.HasPrefix(m.Message, ch.Prefix) {
			args := m.Message[len(ch.Prefix):]
			if embd, ok := handler.Bot.Library.GetItem(strings.ToLower(args)); ok {
				handler.Session.Say(m.Channel, createResponse(embd))
				handler.Bot.Logger.Printf("[CMD] Command %s Successful!", args)
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

//Joining initial twitch channels
func (handler *TwitchHandler) JoinInitialChans() {
	for k := range handler.Bot.Database.Chans {
		//Trying to join non-discord channels
		if IsTChan(k) != "" {
			handler.Session.Join(k)
		}
	}
	//Joining self channel
	handler.Session.Join(handler.Bot.Creds.TName)
}
