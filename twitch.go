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
	h := &TwitchHandler{Bot: bot}
	h.Session = twitch.NewClient(bot.Creds.TName, bot.Creds.TToken)
	h.Session.OnPrivateMessage(h.twitchMessage)
	h.JoinInitialChans()
	return h
}

func (h *TwitchHandler) Start() {
	go h.Session.Connect()
	h.Bot.Logger.Println("[SETUP] Connecting to Twitch")
}

func (h *TwitchHandler) Stop() {
	h.Session.Disconnect()
}

var IsTChan = regexp.MustCompile(`[^0-9]`).FindString

//Handles incoming twitch messages
func (h *TwitchHandler) twitchMessage(m twitch.PrivateMessage) {
	//Checking self channel for join requests
	if strings.EqualFold(m.Channel, h.Bot.Creds.TName) {
		if m.Message == "!biteme" || m.Message == "!setvamp" {
			ch, _ := h.Bot.Database.CreateChan(strings.ToLower(m.User.Name), "!")
			h.Bot.Database.Chans[strings.ToLower(m.User.Name)] = ch
			h.Session.Join(strings.ToLower(m.User.Name))
			h.Session.Say(m.Channel, fmt.Sprintf("Ouch! You've been bitten! I will now respond to commands in %v!", m.User.Name))
		}
		return
	}
	//Regular commands
	if ch, ok := h.Bot.Database.Chans[m.Channel]; ok {
		if strings.HasPrefix(m.Message, ch.Prefix) {
			args := m.Message[len(ch.Prefix):]
			if embd, ok := h.Bot.Library.GetItem(strings.ToLower(args), true); ok {
				h.Session.Say(m.Channel, h.createResponse(embd))
				h.Bot.Logger.Printf("[CMD] Command %s Successful!", args)
				return
			}
		}
		return
	}
}

//Converts library content to Twitch appropriate message
func (h *TwitchHandler) createResponse(content discordgo.MessageEmbed) string {
	var fields string
	for _, embed_f := range content.Fields {
		fields = fields + fmt.Sprintf("%s: %s. ", embed_f.Name, embed_f.Value)
	}
	fields = strings.Replace(fields, "\n", " ", -1)
	result := fmt.Sprintf("%s: %s | %s", content.Title, content.Description, fields)
	result = strings.Replace(result, "||", "", -1)
	if len(result) >= 500 {
		result = result[:496] + "..."
	}
	return result
}

//Joining initial twitch channels
func (h *TwitchHandler) JoinInitialChans() {
	for k := range h.Bot.Database.Chans {
		//Trying to join non-discord channels
		if IsTChan(k) != "" {
			h.Session.Join(k)
		}
	}
	//Joining self channel
	h.Session.Join(h.Bot.Creds.TName)
}
