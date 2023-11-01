package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Krognol/go-wolfram"
	"github.com/joho/godotenv"
	"github.com/shomali11/slacker"
	"github.com/tidwall/gjson"
	witai "github.com/wit-ai/wit-go/v2"
)

func printCommandEvents(analyticsChannel <- chan *slacker.CommandEvent) {
	for event := range analyticsChannel {
		fmt.Println("Command Events")
		fmt.Println(event.Timestamp)
		fmt.Println(event.Command)
		fmt.Println(event.Parameters)
		fmt.Println(event.Event)
		fmt.Println()
	}
}

func main() {
	godotenv.Load(".env")

	bot := slacker.NewClient(os.Getenv("SLACK_BOT_TOKEN"), os.Getenv("SLACK_APP_TOKEN"))

	client := witai.NewClient(os.Getenv("WIT_AI_TOKNE"))
	wolframClient := &wolfram.Client{AppID: os.Getenv("WOLFRAM_APP_ID")}

	go printCommandEvents(bot.CommandEvents())

	bot.Command("query for bot - <message>", &slacker.CommandDefinition{
		Description: "send any question to wolfram",
		Examples: []string{"who is the president of India"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter){
			query := request.Param("message")
			fmt.Println(query)
			msg, err := client.Parse(&witai.MessageRequest{
				Query: query,
			})
			if err != nil {
				log.Printf("failed to parse query: %v\n", err)
			}

			data, err := json.MarshalIndent(msg, "", "    ")
			if err != nil {
				log.Printf("failed to marshal msg: %v\n", err)
			}
			rough := string(data[:])
			value := gjson.Get(rough, "entities.wit$wolfram_search_query:wolfram_search_query.0.value")

			answer := value.String()
			res, err := wolframClient.GetSpokentAnswerQuery(answer, wolfram.Metric, 1000)
			if err != nil {
				log.Printf("failed to get response from wolfram: %v\n", err)
			}
			fmt.Println(res)
			response.Reply(res)
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
