package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Embeder interface {
	Embed() *discordgo.MessageEmbed
}

type Category struct {
	Name    string
	Items   map[string]*DbItem
	Content *discordgo.MessageEmbed
}

func (cat *Category) Embed() *discordgo.MessageEmbed {
	return cat.Content
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

type LibraryHandler struct {
	Bot        *VampBot
	Path       string
	Library    map[string]Embeder
	Categories map[string]Embeder
	Aliases    map[string]string
}

func MakeLibraryHandler(bot *VampBot) *LibraryHandler {
	handler := &LibraryHandler{Path: "library"}
	return handler
}

func (handler *LibraryHandler) LoadLibrary() {
	handler.Library = make(map[string]Embeder)
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
				if item.Metadata.Spoiler || item.Metadata.Beta {
					category.Content.Fields[0].Value += fmt.Sprintf("||%s||, ", item.Metadata.Name)
				} else {
					category.Content.Fields[0].Value += fmt.Sprintf("%s, ", item.Metadata.Name)
				}
			}
			fobj.Close()
		}
		handler.Library[category.Name] = category
		handler.Categories[category.Name] = category
	}
}

func (handler *LibraryHandler) GetItem(args string) (discordgo.MessageEmbed, bool) {
	ret := discordgo.MessageEmbed{}
	return ret, true
}
