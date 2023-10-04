package vampbot

import (
	"fmt"
	"sort"
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

// Handles messages
func (h *DiscordHandler) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	//Owner commands
	if m.Author.ID == h.Bot.Creds.Owner {
		if strings.HasPrefix(m.Content, "!vampstatus") {
			h.StatusCmd(s, m.ChannelID)
			h.Bot.Logger.Printf("[CMD] Sent Bot Status!")
		}
		if strings.HasPrefix(m.Content, "!vamplist") {
			h.ListCmd(s, m.ChannelID)
			h.Bot.Logger.Printf("[CMD] Sent Server List!")
		}
	}
	//Admin commands
	if admin, _ := IsAdmin(s, m); admin || m.Author.ID == h.Bot.Creds.Owner {
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
			} else {
				s.ChannelMessageSend(m.ChannelID, "```Item not found!```")
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

// Checks wether message author has administrator permissions
func IsAdmin(s *discordgo.Session, m *discordgo.MessageCreate) (bool, error) {
	perm, err := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
	if err != nil {
		return false, err
	}
	return perm&discordgo.PermissionAdministrator != 0, nil
}

// Bot status command
func (h *DiscordHandler) StatusCmd(s *discordgo.Session, c string) {
	guilds := s.State.Guilds
	m := discordgo.MessageEmbed{
		Title:       "Vampire Survivors Bot Status",
		Color:       12846604,
		Description: "This bot is free and open source. \nCheck the [GitHub](https://github.com/SHA65536/Vampire-Survivors-Bot)",
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Connected Servers", Value: fmt.Sprint(len(guilds))},
			{Name: "Number Of Library Entries", Value: fmt.Sprint(len(h.Bot.Library.DLibrary))},
		},
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: "https://ucarecdn.com/dc6980a3-c7d4-4164-9b02-276ea7832791/BotIcon.png",
			Text:    "Vampire Bot @ github.com/SHA65536/Vampire-Survivors-Bot",
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://ucarecdn.com/0bb2dc9b-f74e-4d1a-a106-af81331cb5cd/Bat.gif",
		},
	}
	err := SendEmbed(s, c, m)
	if err != nil {
		h.Bot.Logger.Print(err)
	}
}

// Bot status command
func (h *DiscordHandler) ListCmd(s *discordgo.Session, c string) {
	guilds := s.State.Guilds
	sort.Slice(guilds, func(i, j int) bool {
		return guilds[i].MemberCount > guilds[j].MemberCount
	})
	fields := []*discordgo.MessageEmbedField{}
	curf := &discordgo.MessageEmbedField{Name: "Members - Name - ID"}
	informations := []string{}
	for _, g := range guilds {
		last_for_id := g.ID[len(g.ID)-4:]
		informations = append(informations, fmt.Sprintf("%d - %s - %s\n", g.MemberCount, g.Name, last_for_id))
	}
	for _, msg := range informations {
		if len(curf.Value)+len(msg) >= 1024 {
			fields = append(fields, curf)
			curf = &discordgo.MessageEmbedField{Name: "Members - Name - ID"}
		}
		curf.Value += msg
	}
	fields = append(fields, curf)
	m := discordgo.MessageEmbed{
		Title:       "Vampire Survivors Server List",
		Color:       12846604,
		Description: "This bot is free and open source. \nCheck the [GitHub](https://github.com/SHA65536/Vampire-Survivors-Bot)",
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: "https://ucarecdn.com/dc6980a3-c7d4-4164-9b02-276ea7832791/BotIcon.png",
			Text:    "Vampire Bot @ github.com/SHA65536/Vampire-Survivors-Bot",
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://ucarecdn.com/0bb2dc9b-f74e-4d1a-a106-af81331cb5cd/Bat.gif",
		},
	}
	err := SendEmbed(s, c, m)
	if err != nil {
		h.Bot.Logger.Print(err)
	}
}
