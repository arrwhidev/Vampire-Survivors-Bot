package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
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
	Library    map[string]Embeder
	Categories map[string]Embeder
	Aliases    map[string]string
	Fuzzy      []string
	Emotes     map[string]string
}

func MakeLibraryHandler(bot *VampBot) *LibraryHandler {
	handler := &LibraryHandler{Path: "library"}
	handler.LoadLibrary()
	handler.LoadAliases()
	handler.LoadHelp()
	handler.LoadBeta()
	handler.LoadEmotes()
	return handler
}

func (handler *LibraryHandler) LoadLibrary() {
	handler.Library = make(map[string]Embeder)
	handler.Categories = make(map[string]Embeder)
	handler.Fuzzy = make([]string, 0)
	dirs, err := ioutil.ReadDir(handler.Path)
	if err != nil {
		handler.Bot.Logger.Fatal(err)
	}
	for _, dir := range dirs {
		if !dir.IsDir() || len(dir.Name()) <= 3 {
			continue
		}
		category := &Category{Name: dir.Name(), Content: &discordgo.MessageEmbed{}}
		files, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", handler.Path, dir.Name()))
		if err != nil {
			handler.Bot.Logger.Fatal(err)
		}
		for _, file := range files {
			fobj, err := os.Open(fmt.Sprintf("%s/%s/%s", handler.Path, dir.Name(), file.Name()))
			if err != nil {
				handler.Bot.Logger.Fatal(err)
			}
			data, err := ioutil.ReadAll(fobj)
			if err != nil {
				handler.Bot.Logger.Fatal(err)
			}
			if strings.EqualFold(file.Name(), ".category") {
				json.Unmarshal(data, category.Content)
			} else if strings.HasSuffix(file.Name(), ".json") {
				item := &DbItem{}
				json.Unmarshal(data, item)
				handler.Library[item.Metadata.Name] = item
				handler.Fuzzy = append(handler.Fuzzy, item.Metadata.Name)
				if item.Metadata.Spoiler || item.Metadata.Beta {
					category.Content.Fields[0].Value += fmt.Sprintf("||%s||, ", item.Metadata.Name)
				} else {
					category.Content.Fields[0].Value += fmt.Sprintf("%s, ", item.Metadata.Name)
				}
			}
			fobj.Close()
		}
		handler.Fuzzy = append(handler.Fuzzy, category.Name)
		handler.Library[category.Name] = category
		handler.Categories[category.Name] = category
	}
}

func (handler *LibraryHandler) LoadAliases() {
	handler.Aliases = make(map[string]string)
	fobj, err := os.Open(fmt.Sprintf("%s/aliases.json", handler.Path))
	if err != nil {
		handler.Bot.Logger.Fatal(err)
	}
	data, err := ioutil.ReadAll(fobj)
	if err != nil {
		handler.Bot.Logger.Fatal(err)
	}
	json.Unmarshal(data, &handler.Aliases)
	fobj.Close()
}

func (handler *LibraryHandler) LoadHelp() {
	help := &DbItem{}
	fobj, err := os.Open(fmt.Sprintf("%s/help.json", handler.Path))
	if err != nil {
		handler.Bot.Logger.Fatal(err)
	}
	data, err := ioutil.ReadAll(fobj)
	if err != nil {
		handler.Bot.Logger.Fatal(err)
	}
	json.Unmarshal(data, help)
	for name := range handler.Categories {
		help.Content.Fields[1].Value += fmt.Sprintf("%s, ", name)
	}
	help.Content.Fields[1].Value += "beta"
	handler.Library["help"] = help
	fobj.Close()
}

func (handler *LibraryHandler) LoadBeta() {
	beta := &DbItem{}
	fobj, err := os.Open(fmt.Sprintf("%s/beta.json", handler.Path))
	if err != nil {
		handler.Bot.Logger.Fatal(err)
	}
	data, err := ioutil.ReadAll(fobj)
	if err != nil {
		handler.Bot.Logger.Fatal(err)
	}
	json.Unmarshal(data, beta)
	for name, value := range handler.Library {
		if value.Beta() {
			beta.Content.Fields[0].Value += fmt.Sprintf("%s\n", name)
		}
	}
	beta.Content.Fields[0].Value += "||"
	handler.Library["beta"] = beta
	fobj.Close()
}

func (handler *LibraryHandler) LoadEmotes() {
	emotes := make(map[string]string)
	fobj, err := os.Open(fmt.Sprintf("%s/emotes.json", handler.Path))
	if err != nil {
		handler.Bot.Logger.Fatal(err)
	}
	data, err := ioutil.ReadAll(fobj)
	if err != nil {
		handler.Bot.Logger.Fatal(err)
	}
	json.Unmarshal(data, &emotes)
	handler.Emotes = emotes
}

func (handler *LibraryHandler) GetItem(args string) (discordgo.MessageEmbed, bool) {
	if embed, ok := handler.Library[args]; ok {
		return *embed.Embed(), true
	}
	if key, ok := handler.Aliases[args]; ok {
		embed := handler.Library[key]
		return *embed.Embed(), true
	}
	if key := handler.FuzzySearch(args); key != "" {
		embed := *handler.Library[key].Embed()
		embed.Title = fmt.Sprintf("Did you mean: %s?", embed.Title)
		return embed, true
	}
	return discordgo.MessageEmbed{}, false
}

func (handler *LibraryHandler) FuzzySearch(args string) (name string) {
	matches := fuzzy.Find(args, handler.Fuzzy)
	if len(matches) > 0 {
		return matches[0].Str
	}
	return ""
}
