package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"github.com/thteam47/Bot_Slack_Golang/drive"
)

type ServerImpl struct {
	Sockermode *socketmode.Client
	ApiSlack   *slack.Client
}

func (s ServerImpl) sendMess(rw http.ResponseWriter, r *http.Request) {
	user := r.FormValue("user")
	pass := r.FormValue("pass")
	dst := make(chan bool)
	resp := ""
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if user == "admin" && pass == "admin" {
		_, _, err := s.ApiSlack.PostMessage(
			"C036YAGNV8Q",
		)
		if err != nil {
			log.Println(err)
		}

		go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
			// Create a for loop that selects either the context cancellation or the events incomming
			for {
				select {
				// inscase context cancel is called exit the goroutine
				case <-ctx.Done():
					log.Println("Shutting down socketmode listener")
					return
				case event := <-socketClient.Events:
					switch event.Type {
					case socketmode.EventTypeEventsAPI:
						continue
					case socketmode.EventTypeSlashCommand:
						command, ok := event.Data.(slack.SlashCommand)
						if !ok {
							log.Printf("Could not type cast the message to a SlashCommand: %v\n", command)
							continue
						}

						socketClient.Ack(*event.Request)
						isApprove := handleSlashCommand(command, client)
						if isApprove {
							dst <- true
						} else {
							dst <- false
						}

					}
				}

			}
		}(ctx, s.ApiSlack, s.Sockermode)

		check := <-dst
		if check {
			resp = "Login Succes"

		} else {
			resp = "Decline Auth"
		}
		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(resp)

	} else {
		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode("User or Pass incorrect")

	}
}
func (s ServerImpl) sendMessAu() {
	channelID := "C036YAGNV8Q"
	for {
		if s.ApiSlack != nil {
			attachment := slack.Attachment{
				Pretext: "Report: ",
				Text:    "some text",
				Color:   "#36a64f",
				Fields: []slack.AttachmentField{
					{
						Title: "Date",
						Value: time.Now().String(),
					},
				},
			}
			_, _, err := s.ApiSlack.PostMessage(
				channelID,
				slack.MsgOptionAttachments(attachment),
			)
			if err != nil {
				log.Println(err)
			}
			time.Sleep(20 * time.Second)
		}
	}
}

func handleSlashCommand(command slack.SlashCommand, client *slack.Client) bool {
	// We need to switch depending on the command
	switch command.Command {
	case "/approve":
		handleCommand(command, client, true)
		return true
	case "/decline":
		handleCommand(command, client, false)
		return false
	}
	return true
}
func handleCommand(command slack.SlashCommand, client *slack.Client, isApprove bool) {
	resp := ""
	if isApprove {
		resp = "Login Succes"

	} else {
		resp = "Decline Auth"
	}
	attachment := slack.Attachment{}
	attachment.Fields = []slack.AttachmentField{
		{
			Title: resp,
			Value: time.Now().String(),
		},
	}

	attachment.Text = fmt.Sprintf("Response %s", command.Text)
	attachment.Color = "#4af030"

	_, _, err := client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
	if err != nil {
		log.Panic(err)
	}
}
func getlogin(rw http.ResponseWriter, r *http.Request) {
	tmp := template.Must(template.ParseFiles("template.html"))
	tmp.Execute(rw, nil)
}
func main() {

	bot := drive.ConnectBot()
	server := ServerImpl{Sockermode: bot.SocketClient, ApiSlack: bot.BotSlack}
	go bot.SocketClient.Run()
	go server.sendMessAu()
	r := mux.NewRouter()
	r.HandleFunc("/login", getlogin).Methods("GET")
	r.HandleFunc("/login", server.sendMess).Methods("POST")
	http.ListenAndServe(":8080", r)
}
