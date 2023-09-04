package main

import (
	"fmt"
	"os"

	"github.com/slack-go/slack"
)

func main()  {
	os.Setenv("SLACK_BOT_TOKEN", "xoxb-5848665325636-5843323916261-pRMOtpE5mwh7zMspTovM4GUy")
	os.Setenv("CHANNEL_ID", "C05QW6N8FC2")
	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))
	channelArr := []string{os.Getenv("CHANNEL_ID")}
	fileArr := []string{"/home/pirate/workspace/go/go-practice/slack-file-bot/test"}

	for i := 0; i < len(fileArr); i++ {
		params := slack.FileUploadParameters{
			Channels: channelArr,
			File: fileArr[i],
		}

		file, err := api.UploadFile(params)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		fmt.Printf("Name: %s, URL: %s\n", file.Name, file.URL)
	}
}
