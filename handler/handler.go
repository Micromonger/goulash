// Package handler provides an HTTP handler for processing Slack Slash Command
// callbacks. See https://api.slack.com/slash-commands for more information.
package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/action"
	"github.com/pivotalservices/slack"
)

// Handler is an HTTP handler.
type Handler struct {
	api                action.SlackAPI
	slackTeamName      string
	slackUserID        string
	uninvitableDomain  string
	uninvitableMessage string
	auditLogChannelID  string
	clock              clock.Clock
	logger             lager.Logger
}

// New returns a new Handler.
func New(
	api action.SlackAPI,
	slackTeamName string,
	slackUserID string,
	uninvitableDomain string,
	uninvitableMessage string,
	auditLogChannelID string,
	clock clock.Clock,
	logger lager.Logger,
) *Handler {
	return &Handler{
		api:                api,
		slackTeamName:      slackTeamName,
		slackUserID:        slackUserID,
		uninvitableDomain:  uninvitableDomain,
		uninvitableMessage: uninvitableMessage,
		auditLogChannelID:  auditLogChannelID,
		clock:              clock,
		logger:             logger,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	channelID := r.PostFormValue("channel_id")
	channelName := r.PostFormValue("channel_name")
	commanderID := r.PostFormValue("user_id")
	commanderName := r.PostFormValue("user_name")
	text := r.PostFormValue("text")

	if channelID == "" || text == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.logger.Info("started-processing-request", lager.Data{
		"channelID":     channelID,
		"channelName":   channelName,
		"commanderID":   commanderID,
		"commanderName": commanderName,
		"text":          text,
	})

	a := action.New(
		channelID,
		channelName,
		commanderName,
		commanderID,
		text,

		h.api,
		h.slackTeamName,
		h.slackUserID,
		h.uninvitableDomain,
		h.uninvitableMessage,
		h.logger,
	)

	if a, ok := a.(action.GuardedAction); ok {
		checkErr := a.Check()
		if checkErr != nil {
			respondWith(checkErr.Error(), w, h.logger)
			return
		}
	}

	result, err := a.Do()

	if h.auditLogChannelID != "" {
		if a, ok := a.(action.AuditableAction); ok {
			h.postAuditLogEntry(a.AuditMessage(), err)
		}
	}

	if err != nil {
		h.logger.Error("failed-to-perform-request", err)
	}

	respondWith(result, w, h.logger)

	h.logger.Info("finished-processing-request")
}

func (h *Handler) postAuditLogEntry(text string, err error) {
	var outcome string
	if err == nil {
		outcome = "was successful."
	} else {
		outcome = fmt.Sprintf("failed with error: %s", err.Error())
	}

	message := fmt.Sprintf("%s at %s, which %s", text, h.clock.Now().UTC().Round(time.Second), outcome)

	postMessageParameters := slack.NewPostMessageParameters()
	postMessageParameters.AsUser = true
	postMessageParameters.Parse = "full"

	_, _, err = h.api.PostMessage(h.auditLogChannelID, message, postMessageParameters)
	if err != nil {
		h.logger.Error("failed-to-add-audit-log-entry", err)
		return
	}

	h.logger.Info("successfully-added-audit-log-entry")
}

func respondWith(text string, w http.ResponseWriter, logger lager.Logger) {
	_, err := w.Write([]byte(text))
	if err != nil {
		logger.Error("failed-writing-response-body", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
