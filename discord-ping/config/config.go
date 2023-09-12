package config

import (
	"encoding/json"
	"log"
	"os"
)

var (
	Token string
	BotPrefix string

	config *configStruct
)

type configStruct struct {
	Token string `json:"Token"`
	BotPrefix string `json:"BotPrefix"`
}

func ReadConfig() error {
	log.Println("Reading config file...")

	file, err := os.ReadFile("./config.json")
	if err != nil {
		log.Printf("error reading config file: %v", err)
	}

	log.Printf("Config file: %v", string(file))

	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Printf("failed to parse config file: %v", err)
		return err
	}

	Token = config.Token
	BotPrefix = config.BotPrefix

	return nil
}
