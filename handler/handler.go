package handler

import (
	"fmt"
	"net/http"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body := fmt.Sprintf("command='%s' text='%s'", r.PostFormValue("command"), r.PostFormValue("text"))
	w.Write([]byte(body))
}
