// Package handler provides an HTTP handler for processing Slack Slash Command
// callbacks. See https://api.slack.com/slash-commands for more information.
package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/krishicks/slack"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
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

	channelID := r.PostFormValue("channel_id")
	text := r.PostFormValue("text")

	if channelID == "" || text == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	command := text[:strings.IndexByte(text, 0x20)]

	h.logger.Info("started-processing-request", lager.Data{
		"channel_id": channelID,
		"text":       text,
	})

	switch command {
	case "invite-guest":
		form := r.PostForm
		textParams := strings.Split(form["text"][0], " ")
		emailAddress := textParams[1]
		firstName := textParams[2]
		lastName := textParams[3]

		action = inviteGuestAction{
			channelID:    channelID,
			invitingUser: form["user_name"][0],
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

	err := action.Do()
	if err != nil {
		h.logger.Error("failed-to-perform-request", err)
		h.report(channelID, fmt.Sprintf("%s: '%s'", action.FailureMessage(), err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if h.auditLogChannelID != "" {
		h.postAuditLogEntry(action.AuditMessage())
	}

	h.report(channelID, action.SuccessMessage())

	h.logger.Info("finished-processing-request")
}

func (h *Handler) report(channelID string, text string) {
	postMessageParameters := slack.NewPostMessageParameters()
	postMessageParameters.AsUser = true

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

	_, _, err := h.api.PostMessage(h.auditLogChannelID, message, postMessageParameters)

	if err != nil {
		h.logger.Error("failed-processing-request", err)
		return
	}
}
