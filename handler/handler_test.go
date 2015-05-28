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
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/butler"},
				"text":         {"invite-guest user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
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
			Ω(actualChannelID).Should(Equal([]string{"C1234567890"}))
			Ω(actualFirstName).Should(Equal("Tom"))
			Ω(actualLastName).Should(Equal("Smith"))
			Ω(actualEmailAddress).Should(Equal("user@example.com"))
		})

		It("responds to Slack with the result of the command on success", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/butler"},
				"text":         {"invite-guest user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &fakes.FakeSlackAPI{}
			h := handler.New(fakeSlackAPI, "fake-team-name", "", fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)
			Ω(w.Body.String()).Should(Equal("@requesting_user invited Tom Smith (user@example.com) as a guest to 'channel-name'"))
		})

		It("responds to Slack with the result of the command on failure", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/butler"},
				"text":         {"invite-guest user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
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
			Ω(w.Body.String()).Should(Equal("Failed to invite Tom Smith (user@example.com) as a guest to 'channel-name': 'failed to invite user'"))
		})

		It("posts a message to the configured audit log channel on success", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/butler"},
				"text":         {"invite-guest user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}

			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			fakeSlackAPI := &fakes.FakeSlackAPI{}
			w := httptest.NewRecorder()
			h := handler.New(fakeSlackAPI, "fake-team-name", "audit-log-channel-id", fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(1))

			expectedParams := slack.NewPostMessageParameters()
			expectedParams.AsUser = true
			expectedParams.Parse = "full"

			actualChannelID, actualText, actualParams := fakeSlackAPI.PostMessageArgsForCall(0)
			Ω(actualChannelID).Should(Equal("audit-log-channel-id"))
			Ω(actualText).Should(Equal("@requesting_user invited Tom Smith (user@example.com) as a single-channel guest to 'channel-name' (C1234567890) at 2014-01-31 10:59:53 +0000 UTC, which was successful."))
			Ω(actualParams).Should(Equal(expectedParams))
		})

		It("posts a message to the configured audit log channel with the group's real name when it can be found", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {handler.PrivateGroupName},
				"command":      {"/butler"},
				"text":         {"invite-guest user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}

			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			fakeSlackAPI := &fakes.FakeSlackAPI{}
			fakeSlackAPI.GetGroupsReturns([]slack.Group{
				{
					Name:        "unexpected-group-1",
					BaseChannel: slack.BaseChannel{Id: "C1111111111"},
				},
				{
					Name:        "channel-name",
					BaseChannel: slack.BaseChannel{Id: "C1234567890"},
				},
				{
					Name:        "unexpected-group-2",
					BaseChannel: slack.BaseChannel{Id: "C9999999999"},
				},
			}, nil)

			w := httptest.NewRecorder()
			h := handler.New(fakeSlackAPI, "fake-team-name", "audit-log-channel-id", fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(1))

			expectedParams := slack.NewPostMessageParameters()
			expectedParams.AsUser = true
			expectedParams.Parse = "full"

			actualChannelID, actualText, actualParams := fakeSlackAPI.PostMessageArgsForCall(0)
			Ω(actualChannelID).Should(Equal("audit-log-channel-id"))
			Ω(actualText).Should(Equal("@requesting_user invited Tom Smith (user@example.com) as a single-channel guest to 'channel-name' (C1234567890) at 2014-01-31 10:59:53 +0000 UTC, which was successful."))
			Ω(actualParams).Should(Equal(expectedParams))
		})

		It("posts a message to the configured audit log channel on failure", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/butler"},
				"text":         {"invite-guest user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}

			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			fakeSlackAPI := &fakes.FakeSlackAPI{}
			fakeSlackAPI.InviteGuestReturns(errors.New("failed to invite user"))

			w := httptest.NewRecorder()
			h := handler.New(fakeSlackAPI, "fake-team-name", "audit-log-channel-id", fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(1))

			expectedParams := slack.NewPostMessageParameters()
			expectedParams.AsUser = true
			expectedParams.Parse = "full"

			actualChannelID, actualText, actualParams := fakeSlackAPI.PostMessageArgsForCall(0)
			Ω(actualChannelID).Should(Equal("audit-log-channel-id"))
			Ω(actualText).Should(Equal("@requesting_user invited Tom Smith (user@example.com) as a single-channel guest to 'channel-name' (C1234567890) at 2014-01-31 10:59:53 +0000 UTC, which failed with error: failed to invite user"))
			Ω(actualParams).Should(Equal(expectedParams))
		})
	})

	Describe("/butler invite-restricted", func() {
		It("invites a restricted account", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/butler"},
				"text":         {"invite-restricted user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &fakes.FakeSlackAPI{}
			h := handler.New(fakeSlackAPI, "fake-team-name", "", fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.InviteRestrictedCallCount()).Should(Equal(1))

			actualTeamName, actualChannelID, actualFirstName, actualLastName, actualEmailAddress := fakeSlackAPI.InviteRestrictedArgsForCall(0)
			Ω(actualTeamName).Should(Equal("fake-team-name"))
			Ω(actualChannelID).Should(Equal("C1234567890"))
			Ω(actualFirstName).Should(Equal("Tom"))
			Ω(actualLastName).Should(Equal("Smith"))
			Ω(actualEmailAddress).Should(Equal("user@example.com"))
		})

		It("responds to Slack with the result of the command on success", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/butler"},
				"text":         {"invite-restricted user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &fakes.FakeSlackAPI{}
			h := handler.New(fakeSlackAPI, "fake-team-name", "", fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(w.Body.String()).Should(Equal("@requesting_user invited Tom Smith (user@example.com) as a restricted account to 'channel-name'"))
		})

		It("responds to Slack with the result of the command on failure", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/butler"},
				"text":         {"invite-restricted user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			fakeSlackAPI := &fakes.FakeSlackAPI{}
			fakeSlackAPI.InviteRestrictedReturns(errors.New("failed to invite user"))

			w := httptest.NewRecorder()
			h := handler.New(fakeSlackAPI, "fake-team-name", "", fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(w.Body.String()).Should(Equal("Failed to invite Tom Smith (user@example.com) as a restricted account to 'channel-name': 'failed to invite user'"))
		})

		It("posts a message to the configured audit log channel on success", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/butler"},
				"text":         {"invite-restricted user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}

			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			fakeSlackAPI := &fakes.FakeSlackAPI{}
			w := httptest.NewRecorder()
			h := handler.New(fakeSlackAPI, "fake-team-name", "audit-log-channel-id", fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(1))

			expectedParams := slack.NewPostMessageParameters()
			expectedParams.AsUser = true
			expectedParams.Parse = "full"

			actualChannelID, actualText, actualParams := fakeSlackAPI.PostMessageArgsForCall(0)
			Ω(actualChannelID).Should(Equal("audit-log-channel-id"))
			Ω(actualText).Should(Equal("@requesting_user invited Tom Smith (user@example.com) as a restricted account to 'channel-name' (C1234567890) at 2014-01-31 10:59:53 +0000 UTC, which was successful."))
			Ω(actualParams).Should(Equal(expectedParams))
		})

		It("posts a message to the configured audit log channel on failure", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/butler"},
				"text":         {"invite-restricted user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}

			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			fakeSlackAPI := &fakes.FakeSlackAPI{}
			fakeSlackAPI.InviteRestrictedReturns(errors.New("failed to invite user"))

			w := httptest.NewRecorder()
			h := handler.New(fakeSlackAPI, "fake-team-name", "audit-log-channel-id", fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(1))

			expectedParams := slack.NewPostMessageParameters()
			expectedParams.AsUser = true
			expectedParams.Parse = "full"

			actualChannelID, actualText, actualParams := fakeSlackAPI.PostMessageArgsForCall(0)
			Ω(actualChannelID).Should(Equal("audit-log-channel-id"))
			Ω(actualText).Should(Equal("@requesting_user invited Tom Smith (user@example.com) as a restricted account to 'channel-name' (C1234567890) at 2014-01-31 10:59:53 +0000 UTC, which failed with error: failed to invite user"))
			Ω(actualParams).Should(Equal(expectedParams))
		})
	})

	Describe("/butler help", func() {
		It("responds to Slack with the help text", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/butler"},
				"text":         {"help"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &fakes.FakeSlackAPI{}
			h := handler.New(fakeSlackAPI, "fake-team-name", "", fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(w.Body.String()).ShouldNot(BeEmpty())
		})
	})
})
