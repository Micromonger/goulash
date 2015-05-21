package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/krishicks/goulash/handler/fakes"

	"github.com/krishicks/goulash/handler"
	"github.com/krishicks/slack"
	"github.com/pivotal-golang/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler", func() {
	Describe("/inviteGuest", func() {
		It("invites a single channel guest", func() {
			reqBody := strings.NewReader("token=some-token&channel_id=C1234567890&command=%2Finvite-guest&text=user@example.com%20Tom%20Smith")
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &fakes.FakeSlackAPI{}
			h := handler.New(fakeSlackAPI, "fake-team-name", lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.InviteGuestCallCount()).Should(Equal(1))

			actualTeamName, actualChannelID, actualFirstName, actualLastName, actualEmailAddress := fakeSlackAPI.InviteGuestArgsForCall(0)
			Ω(actualTeamName).Should(Equal("fake-team-name"))
			Ω(actualChannelID).Should(Equal("C1234567890"))
			Ω(actualFirstName).Should(Equal("Tom"))
			Ω(actualLastName).Should(Equal("Smith"))
			Ω(actualEmailAddress).Should(Equal("user@example.com"))
		})

		It("posts a message to Slack on success", func() {
			reqBody := strings.NewReader("token=some-token&channel_id=C1234567890&command=%2Finvite-guest&text=user@example.com%20Tom%20Smith")
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &fakes.FakeSlackAPI{}
			h := handler.New(fakeSlackAPI, "fake-team-name", lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(1))

			expectedParams := slack.NewPostMessageParameters()
			expectedParams.Text = "Invited user with email address: user@example.com"

			actualChannelID, _, actualParams := fakeSlackAPI.PostMessageArgsForCall(0)
			Ω(actualChannelID).Should(Equal("C1234567890"))
			Ω(actualParams).Should(Equal(expectedParams))
		})

		It("posts a message to Slack on failure", func() {
			reqBody := strings.NewReader("token=some-token&channel_id=C1234567890&command=%2Finvite-guest&text=user@example.com%20Tom%20Smith")
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			fakeSlackAPI := &fakes.FakeSlackAPI{}
			fakeSlackAPI.InviteGuestReturns(errors.New("failed to invite user"))

			w := httptest.NewRecorder()
			h := handler.New(fakeSlackAPI, "fake-team-name", lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(1))

			expectedParams := slack.NewPostMessageParameters()
			expectedParams.Text = "Failed to invite user with email address: user@example.com, 'failed to invite user'"

			actualChannelID, _, actualParams := fakeSlackAPI.PostMessageArgsForCall(0)
			Ω(actualChannelID).Should(Equal("C1234567890"))
			Ω(actualParams).Should(Equal(expectedParams))
		})
	})
})
