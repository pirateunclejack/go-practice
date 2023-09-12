package main

import (
	"log"

	"github.com/pirateunclejack/go-practice/discord-ping/bot"
	"github.com/pirateunclejack/go-practice/discord-ping/config"
)

func main() {
	err := config.ReadConfig()
	if err != nil {
		log.Printf("failed to read config file: %v", err)
		return
	}

	bot.Start()

	<-make(chan struct{})
	return

}