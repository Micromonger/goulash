// Package handler provides an HTTP handler for processing Slack Slash Command
// callbacks. See https://api.slack.com/slash-commands for more information.
package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/krishicks/slack"
	"github.com/pivotal-golang/lager"
)

// Handler is an HTTP handler.
type Handler struct {
	api           SlackAPI
	slackTeamName string

	logger lager.Logger
}

// New returns a new Handler.
func New(api SlackAPI, slackTeamName string, logger lager.Logger) *Handler {
	return &Handler{
		api:           api,
		slackTeamName: slackTeamName,
		logger:        logger,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	channelID := r.PostFormValue("channel_id")
	command := r.PostFormValue("command")
	text := r.PostFormValue("text")

	h.logger.Info("started-processing-request", lager.Data{
		"channel_id": channelID,
		"text":       text,
	})

	switch command {
	case "/invite-guest":
		err := h.inviteGuest(channelID, text)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

	default:
		w.WriteHeader(http.StatusNotFound)
	}

	h.logger.Info("finished-processing-request")
}

func (h *Handler) inviteGuest(channelID string, text string) error {
	params := strings.Split(text, " ")
	emailAddress := params[0]
	firstName := params[1]
	lastName := params[2]

	err := h.api.InviteGuest(
		h.slackTeamName,
		channelID,
		firstName,
		lastName,
		emailAddress,
	)
	if err != nil {
		h.logger.Error("failed-inviting-single-channel-user", err)
		h.report(channelID, fmt.Sprintf("Failed to invite user with email address: %s, '%s'", emailAddress, err.Error()))

		return err
	}

	h.logger.Info("successfully-invited-single-channel-user")
	h.report(channelID, fmt.Sprintf("Invited user with email address: %s", emailAddress))

	return nil
}

func (h *Handler) report(channelID string, text string) {
	postMessageParameters := slack.NewPostMessageParameters()
	postMessageParameters.Text = text

	_, _, err := h.api.PostMessage(channelID, text, postMessageParameters)

	if err != nil {
		h.logger.Error("failed-processing-request", err)
		return
	}
}
