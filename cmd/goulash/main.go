package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/handler"
	"github.com/pivotalservices/slack"
)

const (
	defaultlistenPort = "8080"
	listenPortVar     = "VCAP_APP_PORT"
)

var (
	listenPort string
	listenAddr string

	slackAPI   *slack.Slack
	timekeeper clock.Clock
	logger     lager.Logger
	c          config.Config
	h          *handler.Handler
)

func init() {
	if listenPort = os.Getenv(listenPortVar); listenPort == "" {
		listenPort = defaultlistenPort
	}
	listenAddr = fmt.Sprintf(":%s", listenPort)

	c = config.NewEnvConfig()
	slackAPI = slack.New(c.SlackAuthToken())
	timekeeper = clock.NewClock()
	logger = lager.NewLogger("handler")
	sink := lager.NewReconfigurableSink(lager.NewWriterSink(os.Stdout, lager.DEBUG), lager.DEBUG)
	logger.RegisterSink(sink)

	h = handler.New(c, slackAPI, timekeeper, logger)
}

func main() {
	if err := http.ListenAndServe(listenAddr, h); err != nil {
		log.Fatal("Failed to start server", err)
	}
}
