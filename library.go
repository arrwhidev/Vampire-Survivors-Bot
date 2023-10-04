package vampbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jinzhu/copier"
	"github.com/sahilm/fuzzy"
)

type Embeder interface {
	Embed() *discordgo.MessageEmbed
	Beta() bool
}

type Category struct {
	Name    string
	Items   map[string]*DbItem
	Content *discordgo.MessageEmbed
}

func (cat *Category) Embed() *discordgo.MessageEmbed {
	return cat.Content
}

func (cat *Category) Beta() bool {
	return false
}

type Metadata struct {
	Name    string `json:"name"`
	Spoiler bool   `json:"spoiler"`
	Beta    bool   `json:"beta"`
}

type DbItem struct {
	Metadata Metadata                `json:"metadata"`
	Content  *discordgo.MessageEmbed `json:"content"`
}

func (item *DbItem) Embed() *discordgo.MessageEmbed {
	return item.Content
}
func (item *DbItem) Beta() bool {
	return item.Metadata.Beta
}

type LibraryHandler struct {
	Bot        *VampBot
	Path       string
	DLibrary   map[string]Embeder
	TLibrary   map[string]Embeder
	Categories map[string]Embeder
	Aliases    map[string]string
	Fuzzy      []string
	Emotes     map[string]string
	EmoteRegex *regexp.Regexp
}

func MakeLibraryHandler(bot *VampBot) *LibraryHandler {
	h := &LibraryHandler{Path: "library"}
	h.EmoteRegex = regexp.MustCompile(`<&([^<>&]*)&>`)
	h.LoadEmotes()
	h.LoadLibrary()
	h.LoadAliases()
	h.LoadHelp()
	h.LoadBeta()
	return h
}

func (h *LibraryHandler) LoadLibrary() {
	h.TLibrary = make(map[string]Embeder)
	h.DLibrary = make(map[string]Embeder)
	h.Categories = make(map[string]Embeder)
	h.Fuzzy = make([]string, 0)
	dirs, err := ioutil.ReadDir(h.Path)
	if err != nil {
		h.Bot.Logger.Fatal(err)
	}
	for _, dir := range dirs {
		if !dir.IsDir() || len(dir.Name()) <= 3 {
			continue
		}
		category := &Category{Name: dir.Name(), Content: &discordgo.MessageEmbed{}}
		files, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", h.Path, dir.Name()))
		if err != nil {
			h.Bot.Logger.Fatal(err)
		}
		for _, file := range files {
			fobj, err := os.Open(fmt.Sprintf("%s/%s/%s", h.Path, dir.Name(), file.Name()))
			if err != nil {
				h.Bot.Logger.Fatal(err)
			}
			data, err := ioutil.ReadAll(fobj)
			if err != nil {
				h.Bot.Logger.Fatal(err)
			}
			if strings.EqualFold(file.Name(), ".category") {
				json.Unmarshal(data, category.Content)
			} else if strings.HasSuffix(file.Name(), ".json") {
				item := &DbItem{}
				json.Unmarshal(data, item)
				h.DLibrary[item.Metadata.Name], h.TLibrary[item.Metadata.Name] = h.ConvertEmotes(item)
				h.Fuzzy = append(h.Fuzzy, item.Metadata.Name)
				addToCategory(category, item)
			}
			fobj.Close()
		}
		h.Fuzzy = append(h.Fuzzy, category.Name)
		item := &DbItem{Content: category.Content}
		h.DLibrary[category.Name], h.TLibrary[category.Name] = h.ConvertEmotes(item)
		h.Categories[category.Name] = category
	}
}

func addToCategory(cat *Category, item *DbItem) {
	var name = fmt.Sprintf("%s, ", item.Metadata.Name)
	if item.Metadata.Spoiler || item.Metadata.Beta {
		name = fmt.Sprintf("||%s||, ", item.Metadata.Name)
	}
	var fieldIdx int
	for fieldIdx < len(cat.Content.Fields) {
		if len(cat.Content.Fields[fieldIdx].Value)+len(name) < 1024 {
			break
		}
		fieldIdx++
	}
	if fieldIdx >= len(cat.Content.Fields) {
		cat.Content.Fields = append(cat.Content.Fields, &discordgo.MessageEmbedField{
			Name:   cat.Content.Fields[0].Name,
			Value:  "",
			Inline: cat.Content.Fields[0].Inline,
		})
	}
	cat.Content.Fields[fieldIdx].Value += name
}

func (h *LibraryHandler) LoadAliases() {
	h.Aliases = make(map[string]string)
	fobj, err := os.Open(fmt.Sprintf("%s/aliases.json", h.Path))
	if err != nil {
		h.Bot.Logger.Fatal(err)
	}
	data, err := ioutil.ReadAll(fobj)
	if err != nil {
		h.Bot.Logger.Fatal(err)
	}
	json.Unmarshal(data, &h.Aliases)
	fobj.Close()
}

func (h *LibraryHandler) LoadHelp() {
	help := &DbItem{}
	fobj, err := os.Open(fmt.Sprintf("%s/help.json", h.Path))
	if err != nil {
		h.Bot.Logger.Fatal(err)
	}
	data, err := ioutil.ReadAll(fobj)
	if err != nil {
		h.Bot.Logger.Fatal(err)
	}
	json.Unmarshal(data, help)
	for name := range h.Categories {
		help.Content.Fields[1].Value += fmt.Sprintf("%s, ", name)
	}
	help.Content.Fields[1].Value += "beta"
	h.DLibrary["help"], h.TLibrary["help"] = h.ConvertEmotes(help)
	fobj.Close()
}

func (h *LibraryHandler) LoadBeta() {
	beta := &DbItem{}
	fobj, err := os.Open(fmt.Sprintf("%s/beta.json", h.Path))
	if err != nil {
		h.Bot.Logger.Fatal(err)
	}
	data, err := ioutil.ReadAll(fobj)
	if err != nil {
		h.Bot.Logger.Fatal(err)
	}
	json.Unmarshal(data, beta)

	dbeta, tbeta := h.ConvertEmotes(beta)

	for name, value := range h.DLibrary {
		if value.Beta() {
			dbeta.Content.Fields[0].Value += fmt.Sprintf("%s\n", name)
		}
	}
	dbeta.Content.Fields[0].Value += "||"
	h.DLibrary["beta"] = dbeta

	for name, value := range h.TLibrary {
		if value.Beta() {
			tbeta.Content.Fields[0].Value += fmt.Sprintf("%s\n", name)
		}
	}
	tbeta.Content.Fields[0].Value += "||"
	h.TLibrary["beta"] = tbeta
	fobj.Close()
}

func (h *LibraryHandler) LoadEmotes() {
	emotes := make(map[string]string)
	fobj, err := os.Open(fmt.Sprintf("%s/emotes.json", h.Path))
	if err != nil {
		h.Bot.Logger.Fatal(err)
	}
	data, err := ioutil.ReadAll(fobj)
	if err != nil {
		h.Bot.Logger.Fatal(err)
	}
	json.Unmarshal(data, &emotes)
	h.Emotes = emotes
}

func (h *LibraryHandler) ConvertEmotes(input *DbItem) (discord *DbItem, twitch *DbItem) {
	discord, twitch = &DbItem{}, &DbItem{}
	copier.Copy(&discord, &input)
	discord.Content.Fields = make([]*discordgo.MessageEmbedField, 0)
	copier.Copy(twitch, input)
	twitch.Content.Fields = make([]*discordgo.MessageEmbedField, 0)

	discord.Content.Description = h.EmoteRegex.ReplaceAllStringFunc(input.Content.Description, h.EmoteDiscord)
	twitch.Content.Description = h.EmoteRegex.ReplaceAllStringFunc(input.Content.Description, h.EmoteTwitch)

	for _, field := range input.Content.Fields {
		dfield := &discordgo.MessageEmbedField{
			Name:   h.EmoteRegex.ReplaceAllStringFunc(field.Name, h.EmoteDiscord),
			Value:  h.EmoteRegex.ReplaceAllStringFunc(field.Value, h.EmoteDiscord),
			Inline: field.Inline,
		}
		tfield := &discordgo.MessageEmbedField{
			Name:   h.EmoteRegex.ReplaceAllStringFunc(field.Name, h.EmoteTwitch),
			Value:  h.EmoteRegex.ReplaceAllStringFunc(field.Value, h.EmoteTwitch),
			Inline: field.Inline,
		}
		discord.Content.Fields = append(discord.Content.Fields, dfield)
		twitch.Content.Fields = append(twitch.Content.Fields, tfield)
	}
	return
}

func (h *LibraryHandler) EmoteDiscord(input string) string {
	stripped := input[2 : len(input)-2]
	if res, ok := h.Emotes[stripped]; ok {
		return res
	}
	return input
}

func (h *LibraryHandler) EmoteTwitch(input string) string {
	return input[2 : len(input)-2]
}

func (h *LibraryHandler) GetItem(args string, twitch bool) (discordgo.MessageEmbed, bool) {
	lib := &h.DLibrary
	if twitch {
		lib = &h.TLibrary
	}
	if embed, ok := (*lib)[args]; ok {
		return *embed.Embed(), true
	}
	if key, ok := h.Aliases[args]; ok {
		embed := (*lib)[key]
		return *embed.Embed(), true
	}
	if key := h.FuzzySearch(args); key != "" {
		embed := *(*lib)[key].Embed()
		embed.Title = fmt.Sprintf("Did you mean: %s?", embed.Title)
		return embed, true
	}
	return discordgo.MessageEmbed{}, false
}

func (h *LibraryHandler) FuzzySearch(args string) (name string) {
	matches := fuzzy.Find(args, h.Fuzzy)
	if len(matches) > 0 {
		return matches[0].Str
	}
	return ""
}
