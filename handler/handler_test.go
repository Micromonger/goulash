package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/pivotalservices/goulash/handler/fakes"

	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/handler"
	"github.com/pivotalservices/slack"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler", func() {
	var (
		fakeClock   *fakeclock.FakeClock
		initialTime time.Time
	)

	BeforeEach(func() {
		initialTime = time.Date(2014, 1, 31, 10, 59, 53, 124235, time.UTC)
		fakeClock = fakeclock.NewFakeClock(initialTime)
	})

	It("returns 400 when given a request with a form not including a channel_id field", func() {
		v := url.Values{
			"token":     {"some-token"},
			"command":   {"/butler"},
			"text":      {"invite-guest user@example.com Tom Smith"},
			"user_name": {"requesting_user"},
		}
		reqBody := strings.NewReader(v.Encode())
		r, err := http.NewRequest("POST", "http://localhost", reqBody)
		Ω(err).ShouldNot(HaveOccurred())

		r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

		w := httptest.NewRecorder()
		fakeSlackAPI := &fakes.FakeSlackAPI{}
		h := handler.New(fakeSlackAPI, "fake-team-name", "", fakeClock, lager.NewLogger("fakelogger"))
		h.ServeHTTP(w, r)

		Ω(w.Code).Should(Equal(http.StatusBadRequest))
	})

	It("returns 400 when given a request with a form not including a text field", func() {
		v := url.Values{
			"token":      {"some-token"},
			"channel_id": {"C1234567890"},
			"command":    {"/butler"},
			"user_name":  {"requesting_user"},
		}
		reqBody := strings.NewReader(v.Encode())
		r, err := http.NewRequest("POST", "http://localhost", reqBody)
		Ω(err).ShouldNot(HaveOccurred())

		r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

		w := httptest.NewRecorder()
		fakeSlackAPI := &fakes.FakeSlackAPI{}
		h := handler.New(fakeSlackAPI, "fake-team-name", "", fakeClock, lager.NewLogger("fakelogger"))
		h.ServeHTTP(w, r)

		Ω(w.Code).Should(Equal(http.StatusBadRequest))
	})

	Describe("/butler invite-guest", func() {
		It("invites a single channel guest", func() {
			v := url.Values{
				"token":      {"some-token"},
				"channel_id": {"C1234567890"},
				"command":    {"/butler"},
				"text":       {"invite-guest user@example.com Tom Smith"},
				"user_name":  {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &fakes.FakeSlackAPI{}
			h := handler.New(fakeSlackAPI, "fake-team-name", "", fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.InviteGuestCallCount()).Should(Equal(1))

			actualTeamName, actualChannelID, actualFirstName, actualLastName, actualEmailAddress := fakeSlackAPI.InviteGuestArgsForCall(0)
			Ω(actualTeamName).Should(Equal("fake-team-name"))
			Ω(actualChannelID).Should(Equal("C1234567890"))
			Ω(actualFirstName).Should(Equal("Tom"))
			Ω(actualLastName).Should(Equal("Smith"))
			Ω(actualEmailAddress).Should(Equal("user@example.com"))
		})

		It("posts a message to the Slack channel that the request came from on success", func() {
			v := url.Values{
				"token":      {"some-token"},
				"channel_id": {"C1234567890"},
				"command":    {"/butler"},
				"text":       {"invite-guest user@example.com Tom Smith"},
				"user_name":  {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &fakes.FakeSlackAPI{}
			h := handler.New(fakeSlackAPI, "fake-team-name", "", fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(1))

			expectedParams := slack.NewPostMessageParameters()
			expectedParams.AsUser = true
			expectedParams.Parse = "full"

			actualChannelID, actualText, actualParams := fakeSlackAPI.PostMessageArgsForCall(0)
			Ω(actualChannelID).Should(Equal("C1234567890"))
			Ω(actualText).Should(Equal("@requesting_user invited Tom Smith (user@example.com) as a guest to this channel"))
			Ω(actualParams).Should(Equal(expectedParams))
		})

		Describe("when inviting the guest fails", func() {
			It("posts a message to the Slack channel that the request came from on failure", func() {
				v := url.Values{
					"token":      {"some-token"},
					"channel_id": {"C1234567890"},
					"command":    {"/butler"},
					"text":       {"invite-guest user@example.com Tom Smith"},
					"user_name":  {"requesting_user"},
				}
				reqBody := strings.NewReader(v.Encode())
				r, err := http.NewRequest("POST", "http://localhost", reqBody)
				Ω(err).ShouldNot(HaveOccurred())

				r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

				fakeSlackAPI := &fakes.FakeSlackAPI{}
				fakeSlackAPI.InviteGuestReturns(errors.New("failed to invite user"))

				w := httptest.NewRecorder()
				h := handler.New(fakeSlackAPI, "fake-team-name", "", fakeClock, lager.NewLogger("fakelogger"))
				h.ServeHTTP(w, r)

				Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(1))

				expectedParams := slack.NewPostMessageParameters()
				expectedParams.AsUser = true
				expectedParams.Parse = "full"

				actualChannelID, actualText, actualParams := fakeSlackAPI.PostMessageArgsForCall(0)
				Ω(actualChannelID).Should(Equal("C1234567890"))
				Ω(actualText).Should(Equal("Failed to invite Tom Smith (user@example.com) as a guest to this channel: 'failed to invite user'"))
				Ω(actualParams).Should(Equal(expectedParams))
			})

			It("posts a message to the Slack audit log channel when an audit log channel is configured", func() {
				v := url.Values{
					"token":      {"some-token"},
					"channel_id": {"C1234567890"},
					"command":    {"/butler"},
					"text":       {"invite-guest user@example.com Tom Smith"},
					"user_name":  {"requesting_user"},
				}

				reqBody := strings.NewReader(v.Encode())
				r, err := http.NewRequest("POST", "http://localhost", reqBody)
				Ω(err).ShouldNot(HaveOccurred())

				r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

				fakeSlackAPI := &fakes.FakeSlackAPI{}

				w := httptest.NewRecorder()
				h := handler.New(fakeSlackAPI, "fake-team-name", "audit-log-channel-id", fakeClock, lager.NewLogger("fakelogger"))
				h.ServeHTTP(w, r)

				Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(2))

				expectedParams := slack.NewPostMessageParameters()
				expectedParams.AsUser = true
				expectedParams.Parse = "full"

				actualChannelID, actualText, actualParams := fakeSlackAPI.PostMessageArgsForCall(0)
				Ω(actualChannelID).Should(Equal("audit-log-channel-id"))
				Ω(actualText).Should(Equal("@requesting_user invited Tom Smith (user@example.com) as a single-channel guest to channel with ID C1234567890 at 2014-01-31 10:59:53 +0000 UTC"))
				Ω(actualParams).Should(Equal(expectedParams))
			})
		})
	})
})
