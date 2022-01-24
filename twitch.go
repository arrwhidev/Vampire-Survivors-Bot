package main

import (
	"fmt"
	"strings"

	"github.com/gempir/go-twitch-irc/v3"
)

//Handles incoming twitch messages
func twitchMessage(m twitch.PrivateMessage) {
	//Checking self channel for join requests
	if strings.ToLower(m.Channel) == strings.ToLower(TName) {
		if m.Message == "!biteme" || m.Message == "!setvamp" {
			ch, _ := CreateChan(strings.ToLower(m.User.Name), "!")
			channels[strings.ToLower(m.User.Name)] = ch
			tclient.Join(strings.ToLower(m.User.Name))
		}
		return
	}
	//Regular commands
	if ch, ok := channels[m.Channel]; ok {
		if strings.HasPrefix(m.Message, ch.Prefix) {
			args := m.Message[len(ch.Prefix):]
			tclient.Say(m.Channel, fmt.Sprintf("Command recognized %s", args))
		}
		return
	}
}
