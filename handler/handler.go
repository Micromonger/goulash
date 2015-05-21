// Package handler provides an HTTP handler for processing Slack Slash Command
// callbacks. See https://api.slack.com/slash-commands for more information.
package handler

import (
	"net/http"
	"os"

	"github.com/nlopes/slack"
	"github.com/pivotal-golang/lager"
)

// Handler is an HTTP handler.
type Handler struct {
	api    SlackAPI
	logger lager.Logger
}

// New returns a new Handler.
func New(api SlackAPI) *Handler {
	logger := lager.NewLogger("handler")
	sink := lager.NewReconfigurableSink(lager.NewWriterSink(os.Stdout, lager.DEBUG), lager.DEBUG)
	logger.RegisterSink(sink)

	return &Handler{
		api:    api,
		logger: logger,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	channelID := r.PostFormValue("channel_id")
	text := r.PostFormValue("text")

	h.logger.Info("started-processing-request", lager.Data{
		"channel_id": channelID,
		"text":       text,
	})

	postMessageParameters := slack.NewPostMessageParameters()
	postMessageParameters.Text = text

	_, _, err := h.api.PostMessage(channelID, text, postMessageParameters)

	if err != nil {
		h.logger.Error("failed-processing-request", err)
		return
	}

	h.logger.Info("finished-processing-request")
}
