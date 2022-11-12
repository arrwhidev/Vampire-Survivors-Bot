# Vampire Turtle
This repository is a community library bot for the game [Vampire Survivors](https://store.steampowered.com/app/1794680/Vampire_Survivors/) written in the Go programming language. Check it out on the official Vampire Survivors [Discord](https://discord.com/invite/vampire-survivors)!

## Discord
The bot is available for invitation to any discord, just [authorize](https://discord.com/api/oauth2/authorize?client_id=761955552091701258&permissions=52224&scope=bot) it.

Once the bot is authorized, use `!setvamp` in the channel you want the bot to respond to commands in. The bot will listen to commands starting with `!` followed by an item you want the info about.

## Installation
Make sure you have [Go](https://go.dev/) installed and in your path.
Clone this repository using:
``` 
git clone https://github.com/SHA65536/Vampire-Survivors-Bot.git
```
Install required libraries using `go get`.

Set the tokens and twitch username using environment variables or create a `.env` file using the following format:
```
TWITCH_TOKEN=<yourtwitchtoken>
DISCORD_TOKEN=<yourdiscordtoken>
TWITCH_NAME=<yourtwitchname>
OWNER_ID=<yourdiscordid>
```

Now run the bot using:
```
go run .
```
## Emotes
In `library/emotes.json` there are emote versions of items and weapons, for Discord this automatically converts to an emotes, but for Twitch it just prints out the item name.

## Contribute
This is a community project, everyone is welcome to suggest and make changes, feel free to open issues / pull requests with your proposed changes.

The data folder is split into categories, each category has a .category file that contains description of it.
Each item has an item.json file containing an embed structure, the name of the file corresponds to the command name. To add aliases for an item, edit aliases.json in the data folder.

## Repository Management
As this is the first repository that I manage, I am looking for feedback/suggestions and potentially someone to help me manage this.
