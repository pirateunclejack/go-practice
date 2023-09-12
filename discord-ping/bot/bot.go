package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/pirateunclejack/go-practice/discord-ping/config"
)

var BotID string
var goBot *discordgo.Session

func Start() {
	goBot, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		log.Printf("failed to new discord session: %v", err)
		return
	}

	u, err := goBot.User("@me")
	if err != nil {
		log.Printf("failed to get go bot information: %v", err)
		return
	}

	BotID = u.ID
	
	goBot.AddHandler(messageHandler)

	err = goBot.Open()

	if err != nil {
		log.Printf("failed to open go bot session: %v", err)
		return
	}

	log.Println("Bot is running")
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate){
	if m.Author.ID == BotID{
		return
	}

	if m.Content == "ping" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "pong")
	}
}
