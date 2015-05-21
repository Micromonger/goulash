// Package handler provides an HTTP handler for processing Slack Slash Command
// callbacks. See https://api.slack.com/slash-commands for more information.
package handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/pivotal-golang/lager"
)

// Handler is an HTTP handler.
type Handler struct {
	logger lager.Logger
}

// New returns a new Handler.
func New() *Handler {
	logger := lager.NewLogger("handler")
	sink := lager.NewReconfigurableSink(lager.NewWriterSink(os.Stdout, lager.DEBUG), lager.DEBUG)
	logger.RegisterSink(sink)

	return &Handler{
		logger: logger,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("started-processing-request")

	body := fmt.Sprintf("command='%s' text='%s'", r.PostFormValue("command"), r.PostFormValue("text"))
	_, err := w.Write([]byte(body))
	if err != nil {
		h.logger.Error("failed-writing-response-body", err)
	}

	h.logger.Info("finished-processing-request")
}
