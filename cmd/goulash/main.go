package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/krishicks/goulash/handler"
	"github.com/krishicks/slack"
	"github.com/pivotal-golang/lager"
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
	logger   lager.Logger
	port     string
)

func init() {
	if port = os.Getenv(portVar); port == "" {
		port = defaultPort
	}
	slackAPI = slack.New(os.Getenv(tokenVar))
	address = fmt.Sprintf(":%s", port)
	logger = lager.NewLogger("handler")
	sink := lager.NewReconfigurableSink(lager.NewWriterSink(os.Stdout, lager.DEBUG), lager.DEBUG)
	logger.RegisterSink(sink)

	h = handler.New(slackAPI, logger)
}

func main() {
	if err := http.ListenAndServe(address, h); err != nil {
		log.Fatal("Failed to start server", err)
	}
}
