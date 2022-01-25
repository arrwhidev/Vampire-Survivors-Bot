# Vampire Turtle
This repository is a community library bot for the game [Vampire Survivors](https://store.steampowered.com/app/1794680/Vampire_Survivors/) written in the Go programming language.

## Discord
The bot is available for invitation to any discord, just [authorize](https://discord.com/api/oauth2/authorize?client_id=761955552091701258&permissions=52224&scope=bot) it.

Once the bot is authorized, use `!vampset` in the channel you want the bot to respond to commands in. The bot will listen to commands starting with `!` followed by an item you want the info about.

## Twitch
The bot is also available for twitch, To addit to your channel, go to the bot's [Twitch channel](https://www.twitch.tv/turtlevampire) and type `!vampset` to add it to your channel. Similarly to discord, the bot will listen to commands starting with `!` followed by an item you want the info about.

## Installation
Make sure you have [Go](https://go.dev/) installed and in your path.
Clone this repository using:
``` 
git clone https://github.com/SHA65536/Vampire-Survivors-Bot.git
```
Install required libraries using `go get`.

Now run the bot using:
```
go run . -t <twitch-oauth> -d <discord-token> -n <twitch-name>
```
Alternatively create a `config.ini` file with this structure:
```json
{"ttoken":"twitch-oauth","dtoken":"discord-token","tname":"twitch-name"}
```

## Contribute
This is a community project, everyone is welcome to suggest and make changes, feel free to open issues / pull requests with your proposed changes.

## Repository Managment
As this is the first repository that I manage, I am looking for feedback / suggestions and potentially someone to help me manage this.