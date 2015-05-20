package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/krishicks/goulash/handler"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler", func() {
	It("responds with the POSTed command and text", func() {
		reqBody := strings.NewReader("token=some-token&command=%2Fthe-command&text=the+text")
		r, err := http.NewRequest("POST", "http://localhost", reqBody)
		Ω(err).ShouldNot(HaveOccurred())

		r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

		w := httptest.NewRecorder()
		h := handler.New()
		h.ServeHTTP(w, r)

		Ω(w.Body.String()).Should(Equal("command='/the-command' text='the text'"))
	})
})
