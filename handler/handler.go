// Package handler provides an HTTP handler for processing Slack Slash Command
// callbacks. See https://api.slack.com/slash-commands for more information.
package handler

import (
	"fmt"
	"net/http"
)

// Handler is an HTTP handler.
type Handler struct{}

// New returns a new Handler.
func New() *Handler {
	return &Handler{}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body := fmt.Sprintf("command='%s' text='%s'", r.PostFormValue("command"), r.PostFormValue("text"))
	w.Write([]byte(body))
}
