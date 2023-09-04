package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/shomali11/slacker/v2"
)

func main() {
	os.Setenv("SLACK_BOT_TOKEN","xoxb-5848665325636-5858939604609-kgOcqukaf1ivm3GJDTxoV1kS")
	os.Setenv("SLACK_APP_TOKEN","xapp-1-A05QTB2BKEH-5843378298837-fcb4bd6cdff47008ebfce3468f79c9a54cbbc1087d2cf3920e84c0a8ba772ba7")

	bot := slacker.NewClient(os.Getenv("SLACK_BOT_TOKEN"), os.Getenv("SLACK_APP_TOKEN"), slacker.WithBotMode(slacker.BotModeIgnoreApp))


	bot.AddCommand(&slacker.CommandDefinition{
		Description: "yob calculator",
		Examples: []string{"my yob is 2020"},
		Command: "my yob is <year>",
		Handler: func(ctx *slacker.CommandContext) {
			year := ctx.Request().Param("year")
			yob, err := strconv.Atoi(year)
			if err != nil {
				fmt.Printf("error: %v", err)
			}
			age := time.Now().Year() - yob
			r := fmt.Sprintf("age is %d", age)
			fmt.Println(ctx.Event())
			ctx.Response().Reply(r)
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bot.Listen(ctx)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
