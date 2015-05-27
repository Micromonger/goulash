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

	switch command {
	case "help":
		_, err = w.Write([]byte(helpText()))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return

	case "invite-guest":
		emailAddress := commandParams[0]
		firstName := commandParams[1]
		lastName := commandParams[2]

		action = inviteGuestAction{
			channel:      channel,
			invitingUser: commanderName,
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
			invitingUser: commanderName,
			emailAddress: emailAddress,
			firstName:    firstName,
			lastName:     lastName,

			api:           h.api,
			slackTeamName: h.slackTeamName,
			logger:        h.logger,
		}

	default:
		_, err = w.Write(helpText())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	err = action.Do()

	if h.auditLogChannelID != "" {
		h.postAuditLogEntry(action.AuditMessage(), err)
	}

	if err != nil {
		h.logger.Error("failed-to-perform-request", err)
		_, err = w.Write([]byte(fmt.Sprintf("%s: '%s'", action.FailureMessage(), err.Error())))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	_, err = w.Write([]byte(action.SuccessMessage()))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

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

func helpText() []byte {
	return []byte("*USAGE*\n" +
		"`/butler [command] [args]`\n" +
		"\n" +
		"*COMMANDS*\n" +
		"\n" +
		"`invite-guest <email> <firstname> <lastname>`\n" +
		"_Invite a Single-Channel Guest to the current channel/group_\n" +
		"\n" +
		"`invite-restricted <email> <firstname> <lastname>`\n" +
		"_Invite a Restricted Account to the current channel/group_\n")
}
