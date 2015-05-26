package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/handler"
	"github.com/pivotalservices/slack"
)

const (
	defaultlistenPort = "8080"

	listenPortVar        = "VCAP_APP_PORT"
	tokenVar             = "SLACK_AUTH_TOKEN"
	teamNameVar          = "SLACK_TEAM_NAME"
	auditLogChannelIDVar = "SLACK_AUDIT_LOG_CHANNEL_ID"
)

var (
	listenPort string
	listenAddr string

	slackAPI          *slack.Slack
	slackTeamName     string
	auditLogChannelID string

	timekeeper clock.Clock
	logger     lager.Logger
	h          *handler.Handler
)

func init() {
	if listenPort = os.Getenv(listenPortVar); listenPort == "" {
		listenPort = defaultlistenPort
	}
	listenAddr = fmt.Sprintf(":%s", listenPort)

	slackAPI = slack.New(os.Getenv(tokenVar))
	slackTeamName = os.Getenv(teamNameVar)
	auditLogChannelID = os.Getenv(auditLogChannelIDVar)

	timekeeper = clock.NewClock()
	logger = lager.NewLogger("handler")
	sink := lager.NewReconfigurableSink(lager.NewWriterSink(os.Stdout, lager.DEBUG), lager.DEBUG)
	logger.RegisterSink(sink)

	h = handler.New(slackAPI, slackTeamName, auditLogChannelID, timekeeper, logger)
}

func main() {
	if err := http.ListenAndServe(listenAddr, h); err != nil {
		log.Fatal("Failed to start server", err)
	}
}
