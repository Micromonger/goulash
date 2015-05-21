package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/krishicks/goulash/handler"
	"github.com/krishicks/goulash/handler/fakes"
	"github.com/krishicks/slack"
	"github.com/pivotal-golang/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler", func() {
	It("posts a message to Slack with the POSTed command and text", func() {
		reqBody := strings.NewReader("token=some-token&channel_id=C1234567890&command=%2Fthe-command&text=the+text")
		r, err := http.NewRequest("POST", "http://localhost", reqBody)
		Ω(err).ShouldNot(HaveOccurred())

		r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

		w := httptest.NewRecorder()
		fakeSlackAPI := &fakes.FakeSlackAPI{}
		h := handler.New(fakeSlackAPI, lager.NewLogger("fakelogger"))
		h.ServeHTTP(w, r)

		Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(1))

		expectedParams := slack.NewPostMessageParameters()
		expectedParams.Text = "the text"

		actualChannelID, _, actualParams := fakeSlackAPI.PostMessageArgsForCall(0)
		Ω(actualChannelID).Should(Equal("C1234567890"))
		Ω(actualParams).Should(Equal(expectedParams))
	})
})
