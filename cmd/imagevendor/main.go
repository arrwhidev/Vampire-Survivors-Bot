package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"vampbot"

	"github.com/joho/godotenv"
)

const imagepath = "imagevendor"

func main() {
	godotenv.Load()
	bot := vampbot.MakeVampBot(
		vampbot.Config{
			DToken: os.Getenv("DISCORD_TOKEN"),
			TToken: os.Getenv("TWITCH_TOKEN"),
			TName:  os.Getenv("TWITCH_NAME"),
			Owner:  os.Getenv("OWNER_ID"),
			Prefix: os.Getenv("PREFIX"),
		})
	var i int
	for _, v := range bot.Library.DLibrary {
		i++
		thumb := v.Embed().Thumbnail
		if thumb == nil {
			continue
		}
		filename := imagepath + "/" + thumb.URL[8:]
		split := strings.Split(filename, "/")
		dir := strings.Join(split[:len(split)-1], "/")
		fmt.Println(i, len(bot.Library.DLibrary), os.MkdirAll(dir, 0777), downloadFile(thumb.URL, filename))
	}
}

func downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}
