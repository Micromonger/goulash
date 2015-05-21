package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/krishicks/goulash/handler"
	"github.com/nlopes/slack"
)

const (
	defaultPort = "8080"

	portVar  = "VCAP_APP_PORT"
	tokenVar = "SLACK_AUTH_TOKEN"
)

var (
	address  string
	slackAPI *slack.Slack
	h        *handler.Handler
	port     string
)

func init() {
	if port = os.Getenv(portVar); port == "" {
		port = defaultPort
	}
	slackAPI = slack.New(os.Getenv(tokenVar))
	address = fmt.Sprintf(":%s", port)
	h = handler.New(slackAPI)
}

func main() {
	if err := http.ListenAndServe(address, h); err != nil {
		log.Fatal("Failed to start server", err)
	}
}
