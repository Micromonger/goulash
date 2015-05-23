package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/krishicks/slack"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/handler"
)

const (
	defaultPort = "8080"

	portVar              = "VCAP_APP_PORT"
	tokenVar             = "SLACK_AUTH_TOKEN"
	teamNameVar          = "SLACK_TEAM_NAME"
	auditLogChannelIDVar = "SLACK_AUDIT_LOG_CHANNEL_ID"
)

var (
	address           string
	slackAPI          *slack.Slack
	slackTeamName     string
	auditLogChannelID string
	h                 *handler.Handler
	logger            lager.Logger
	timekeeper        clock.Clock
	port              string
)

func init() {
	if port = os.Getenv(portVar); port == "" {
		port = defaultPort
	}
	slackAPI = slack.New(os.Getenv(tokenVar))
	slackTeamName = os.Getenv(teamNameVar)
	auditLogChannelID = os.Getenv(auditLogChannelIDVar)
	address = fmt.Sprintf(":%s", port)
	logger = lager.NewLogger("handler")
	sink := lager.NewReconfigurableSink(lager.NewWriterSink(os.Stdout, lager.DEBUG), lager.DEBUG)
	timekeeper = clock.NewClock()
	logger.RegisterSink(sink)

	h = handler.New(slackAPI, slackTeamName, auditLogChannelID, timekeeper, logger)
}

func main() {
	if err := http.ListenAndServe(address, h); err != nil {
		log.Fatal("Failed to start server", err)
	}
}
