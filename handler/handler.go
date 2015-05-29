// Package handler provides an HTTP handler for processing Slack Slash Command
// callbacks. See https://api.slack.com/slash-commands for more information.
package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/slack"
)

const (
	// PrivateGroupName holds the name Slack provides for a Slash Command sent
	// from a group which is private.
	PrivateGroupName = "privategroup"
)

// Handler is an HTTP handler.
type Handler struct {
	api                SlackAPI
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
	api SlackAPI,
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
	channel, commanderName, commanderID, command, commandParams, err := params(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.logger.Info("started-processing-request", lager.Data{
		"channelID":     channel.ID,
		"channelName":   channel.Name(h.api),
		"commanderName": commanderName,
		"commanderID":   commanderID,
		"command":       command,
		"commandParams": commandParams,
	})

	action := NewAction(
		channel,
		commanderName,
		command,
		commandParams,
		channel.Name(h.api),

		h.api,
		h.slackTeamName,
		h.slackUserID,
		h.uninvitableDomain,
		h.uninvitableMessage,
		h.logger,
	)

	if action, ok := action.(GuardedAction); ok {
		checkErr := action.Check()
		if checkErr != nil {
			respondWith(checkErr.Error(), w, h.logger)
			return
		}
	}

	result, err := action.Do()

	if h.auditLogChannelID != "" {
		if action, ok := action.(AuditableAction); ok {
			h.postAuditLogEntry(action.AuditMessage(), err)
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

func params(r *http.Request) (*Channel, string, string, string, []string, error) {
	channelID := r.PostFormValue("channel_id")
	text := r.PostFormValue("text")

	if channelID == "" || text == "" {
		return &Channel{}, "", "", "", []string{}, errors.New("Missing required attributes")
	}

	channel := &Channel{
		RawName: r.PostFormValue("channel_name"),
		ID:      channelID,
	}
	commanderName := r.PostFormValue("user_name")
	commanderID := r.PostFormValue("user_id")

	var command string
	var commandParams []string
	if commandSep := strings.IndexByte(text, 0x20); commandSep > 0 {
		command = text[:commandSep]
		commandParams = strings.Split(text[commandSep+1:], " ")
	} else {
		command = text
	}

	return channel, commanderName, commanderID, command, commandParams, nil
}

func respondWith(text string, w http.ResponseWriter, logger lager.Logger) {
	_, err := w.Write([]byte(text))
	if err != nil {
		logger.Error("failed-writing-response-body", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
