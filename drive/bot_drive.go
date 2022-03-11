package drive

import (
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type BotDb struct {
	BotSlack     *slack.Client
	SocketClient *socketmode.Client
}

var Bot = &BotDb{}

func ConnectBot() *BotDb {
	os.Setenv("SLACK_BOT_TOKEN", "token bot")
	os.Setenv("SLACK_APP_TOKEN", "toekn app")
	appToken := os.Getenv("SLACK_APP_TOKEN")
	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"), slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))
	socketClient := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)
	Bot.BotSlack = api
	Bot.SocketClient = socketClient

	return Bot
}
