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

// Handler is an HTTP handler.
type Handler struct {
	api               SlackAPI
	slackTeamName     string
	auditLogChannelID string

	clock  clock.Clock
	logger lager.Logger
}

// New returns a new Handler.
func New(
	api SlackAPI,
	slackTeamName string,
	auditLogChannelID string,
	clock clock.Clock,
	logger lager.Logger,
) *Handler {
	return &Handler{
		api:               api,
		slackTeamName:     slackTeamName,
		auditLogChannelID: auditLogChannelID,
		clock:             clock,
		logger:            logger,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var action Action

	channel, commander, command, commandParams, err := params(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.logger.Info("started-processing-request", lager.Data{
		"channel_id": channel.ID,
		"text":       fmt.Sprintf("%s %s", command, commandParams),
	})

	switch command {
	case "invite-guest":
		emailAddress := commandParams[0]
		firstName := commandParams[1]
		lastName := commandParams[2]

		action = inviteGuestAction{
			channel:      channel,
			invitingUser: commander,
			emailAddress: emailAddress,
			firstName:    firstName,
			lastName:     lastName,

			api:           h.api,
			slackTeamName: h.slackTeamName,
			logger:        h.logger,
		}

	case "invite-restricted":
		emailAddress := commandParams[0]
		firstName := commandParams[1]
		lastName := commandParams[2]

		action = inviteRestrictedAction{
			channel:      channel,
			invitingUser: commander,
			emailAddress: emailAddress,
			firstName:    firstName,
			lastName:     lastName,

			api:           h.api,
			slackTeamName: h.slackTeamName,
			logger:        h.logger,
		}

	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if h.auditLogChannelID != "" {
		h.postAuditLogEntry(action.AuditMessage())
	}

	err = action.Do()
	if err != nil {
		h.logger.Error("failed-to-perform-request", err)
		h.report(channel.ID, fmt.Sprintf("%s: '%s'", action.FailureMessage(), err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.report(channel.ID, action.SuccessMessage())

	h.logger.Info("finished-processing-request")
}

func (h *Handler) report(channelID string, text string) {
	postMessageParameters := slack.NewPostMessageParameters()
	postMessageParameters.AsUser = true
	postMessageParameters.Parse = "full"

	_, _, err := h.api.PostMessage(channelID, text, postMessageParameters)
	if err != nil {
		h.logger.Error("failed-to-report-message", err)
	}

	h.logger.Info("successfully-reported-message")
}

func (h *Handler) postAuditLogEntry(text string) {
	message := fmt.Sprintf("%s at %s", text, h.clock.Now().UTC().Round(time.Second))

	postMessageParameters := slack.NewPostMessageParameters()
	postMessageParameters.AsUser = true
	postMessageParameters.Parse = "full"

	_, _, err := h.api.PostMessage(h.auditLogChannelID, message, postMessageParameters)
	if err != nil {
		h.logger.Error("failed-to-add-audit-log-entry", err)
		return
	}

	h.logger.Info("successfully-added-audit-log-entry")
}

func params(r *http.Request) (Channel, string, string, []string, error) {
	channelID := r.PostFormValue("channel_id")
	text := r.PostFormValue("text")

	if channelID == "" || text == "" {
		return Channel{}, "", "", []string{}, errors.New("Missing required attributes")
	}

	channel := Channel{
		RawName: r.PostFormValue("channel_name"),
		ID:      channelID,
	}
	commander := r.PostFormValue("user_name")

	var command string
	var commandParams []string
	if commandSep := strings.IndexByte(text, 0x20); commandSep > 0 {
		command = text[:commandSep]
		commandParams = strings.Split(text[commandSep+1:], " ")
	} else {
		command = text
	}

	return channel, commander, command, commandParams, nil
}
