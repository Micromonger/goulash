package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/handler"
	"github.com/pivotalservices/goulash/slackapi"
	"github.com/pivotalservices/goulash/slackapi/slackapifakes"
	"github.com/pivotalservices/slack"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler", func() {
	var (
		c           config.Config
		fakeClock   *fakeclock.FakeClock
		initialTime time.Time
	)

	BeforeEach(func() {
		initialTime = time.Date(2014, 1, 31, 10, 59, 53, 124235, time.UTC)
		fakeClock = fakeclock.NewFakeClock(initialTime)
		c = config.NewLocalConfig(
			"fake-slack-auth-token",
			"/slack-slash-command",
			"slack-team-name",
			"slack-user-id",
			"",
			"uninvitable-domain.com",
			"uninvitable-domain-message",
		)
	})

	It("returns 400 when given a request with a form not including a channel_id field", func() {
		v := url.Values{
			"token":     {"some-token"},
			"command":   {"/slack-slash-command"},
			"text":      {"invite-guest user@example.com Tom Smith"},
			"user_name": {"requesting_user"},
		}
		reqBody := strings.NewReader(v.Encode())
		r, err := http.NewRequest("POST", "http://localhost", reqBody)
		Ω(err).ShouldNot(HaveOccurred())

		r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

		w := httptest.NewRecorder()
		fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
		h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
		h.ServeHTTP(w, r)

		Ω(w.Code).Should(Equal(http.StatusBadRequest))
	})

	It("returns 400 when given a request with a form not including a text field", func() {
		v := url.Values{
			"token":      {"some-token"},
			"channel_id": {"C1234567890"},
			"command":    {"/slack-slash-command"},
			"user_name":  {"requesting_user"},
		}
		reqBody := strings.NewReader(v.Encode())
		r, err := http.NewRequest("POST", "http://localhost", reqBody)
		Ω(err).ShouldNot(HaveOccurred())

		r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

		w := httptest.NewRecorder()
		fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
		h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
		h.ServeHTTP(w, r)

		Ω(w.Code).Should(Equal(http.StatusBadRequest))
	})

	Describe("invite-guest", func() {
		It("invites a single channel guest", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"invite-guest user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.InviteGuestCallCount()).Should(Equal(1))

			actualTeamName, actualChannelID, actualFirstName, actualLastName, actualEmailAddress := fakeSlackAPI.InviteGuestArgsForCall(0)
			Ω(actualTeamName).Should(Equal("slack-team-name"))
			Ω(actualChannelID).Should(Equal("C1234567890"))
			Ω(actualFirstName).Should(Equal("Tom"))
			Ω(actualLastName).Should(Equal("Smith"))
			Ω(actualEmailAddress).Should(Equal("user@example.com"))
		})

		It("invites a single channel guest when first/last name are missing", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"invite-guest user@example.com"}, // first/last names are missing
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.InviteGuestCallCount()).Should(Equal(1))

			actualTeamName, actualChannelID, actualFirstName, actualLastName, actualEmailAddress := fakeSlackAPI.InviteGuestArgsForCall(0)
			Ω(actualTeamName).Should(Equal("slack-team-name"))
			Ω(actualChannelID).Should(Equal("C1234567890"))
			Ω(actualFirstName).Should(BeEmpty())
			Ω(actualLastName).Should(BeEmpty())
			Ω(actualEmailAddress).Should(Equal("user@example.com"))
		})

		It("responds to Slack with the result of the command on success", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"invite-guest user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)
			Ω(w.Body.String()).Should(Equal("Successfully invited Tom Smith (user@example.com) as a single-channel guest to 'channel-name'"))
		})

		It("responds to Slack with the result of the command on failure", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"invite-guest user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			fakeSlackAPI.InviteGuestReturns(errors.New("failed to invite user"))

			w := httptest.NewRecorder()
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)
			Ω(w.Body.String()).Should(Equal("Failed to invite Tom Smith (user@example.com) as a single-channel guest to 'channel-name': failed to invite user"))
		})

		It("responds to Slack when it isn't a member of the private group", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {slackapi.PrivateGroupName},
				"command":      {"/slack-slash-command"},
				"text":         {"invite-guest user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			fakeSlackAPI.GetGroupsReturns([]slack.Group{
				newGroup("unexpected-group-1", "C1111111111"),
				newGroup("unexpected-group-2", "C9999999999"),
			}, nil)

			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)
			Ω(w.Body.String()).Should(Equal("<@slack-user-id> can only invite people to channels or private groups it is a member of. You can invite <@slack-user-id> by typing `/invite @slack-user-id` from the channel or private group you would like <@slack-user-id> to invite people to."))
		})

		It("responds to Slack when an email with an uninvitable domain is invited", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {slackapi.PrivateGroupName},
				"command":      {"/slack-slash-command"},
				"text":         {"invite-guest user@uninvitable-domain.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(w.Body.String()).Should(Equal("Users for the 'uninvitable-domain.com' domain are unable to be invited through /slack-slash-command. uninvitable-domain-message"))
		})

		It("posts a message to the configured audit log channel on success", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"invite-guest user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}

			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			w := httptest.NewRecorder()
			c = config.NewLocalConfig(
				"fake-slack-auth-token",
				"/slack-slash-command",
				"slack-team-name",
				"slack-user-id",
				"audit-log-channel-id",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
			)
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
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
				"channel_name": {slackapi.PrivateGroupName},
				"command":      {"/slack-slash-command"},
				"text":         {"invite-guest user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}

			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			fakeSlackAPI.GetGroupsReturns([]slack.Group{
				newGroup("unexpected-group-1", "C1111111111"),
				newGroup("channel-name", "C1234567890"),
				newGroup("unexpected-group-2", "C9999999999"),
			}, nil)

			w := httptest.NewRecorder()
			c = config.NewLocalConfig(
				"fake-slack-auth-token",
				"/slack-slash-command",
				"slack-team-name",
				"slack-user-id",
				"audit-log-channel-id",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
			)
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
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
				"command":      {"/slack-slash-command"},
				"text":         {"invite-guest user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}

			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			fakeSlackAPI.InviteGuestReturns(errors.New("failed to invite user"))

			w := httptest.NewRecorder()
			c = config.NewLocalConfig(
				"fake-slack-auth-token",
				"/slack-slash-command",
				"slack-team-name",
				"slack-user-id",
				"audit-log-channel-id",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
			)
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
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

	Describe("invite-restricted", func() {
		It("invites a restricted account", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"invite-restricted user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.InviteRestrictedCallCount()).Should(Equal(1))

			actualTeamName, actualChannelID, actualFirstName, actualLastName, actualEmailAddress := fakeSlackAPI.InviteRestrictedArgsForCall(0)
			Ω(actualTeamName).Should(Equal("slack-team-name"))
			Ω(actualChannelID).Should(Equal("C1234567890"))
			Ω(actualFirstName).Should(Equal("Tom"))
			Ω(actualLastName).Should(Equal("Smith"))
			Ω(actualEmailAddress).Should(Equal("user@example.com"))
		})

		It("responds to Slack when an email with an uninvitable domain is invited", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {slackapi.PrivateGroupName},
				"command":      {"/slack-slash-command"},
				"text":         {"invite-restricted user@uninvitable-domain.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			w := httptest.NewRecorder()
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(w.Body.String()).Should(Equal("Users for the 'uninvitable-domain.com' domain are unable to be invited through /slack-slash-command. uninvitable-domain-message"))
		})

		It("invites a restricted account when first/last name are missing", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"invite-restricted user@example.com"}, // first/last names are missing
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.InviteRestrictedCallCount()).Should(Equal(1))

			actualTeamName, actualChannelID, actualFirstName, actualLastName, actualEmailAddress := fakeSlackAPI.InviteRestrictedArgsForCall(0)
			Ω(actualTeamName).Should(Equal("slack-team-name"))
			Ω(actualChannelID).Should(Equal("C1234567890"))
			Ω(actualFirstName).Should(BeEmpty())
			Ω(actualLastName).Should(BeEmpty())
			Ω(actualEmailAddress).Should(Equal("user@example.com"))
		})

		It("responds to Slack when it isn't a member of a private group", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {slackapi.PrivateGroupName},
				"command":      {"/slack-slash-command"},
				"text":         {"invite-restricted user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			fakeSlackAPI.GetGroupsReturns([]slack.Group{
				newGroup("unexpected-group-1", "C1111111111"),
				newGroup("unexpected-group-2", "C9999999999"),
			}, nil)

			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)
			Ω(w.Body.String()).Should(Equal("<@slack-user-id> can only invite people to channels or private groups it is a member of. You can invite <@slack-user-id> by typing `/invite @slack-user-id` from the channel or private group you would like <@slack-user-id> to invite people to."))
		})

		It("responds to Slack with the result of the command on success", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"invite-restricted user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(w.Body.String()).Should(Equal("Successfully invited Tom Smith (user@example.com) as a restricted account to 'channel-name'"))
		})

		It("responds to Slack with the result of the command on failure", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"invite-restricted user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			fakeSlackAPI.InviteRestrictedReturns(errors.New("failed to invite user"))

			w := httptest.NewRecorder()
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(w.Body.String()).Should(Equal("Failed to invite Tom Smith (user@example.com) as a restricted account to 'channel-name': failed to invite user"))
		})

		It("posts a message to the configured audit log channel on success", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"invite-restricted user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}

			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			w := httptest.NewRecorder()
			c = config.NewLocalConfig(
				"fake-slack-auth-token",
				"/slack-slash-command",
				"slack-team-name",
				"slack-user-id",
				"audit-log-channel-id",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
			)
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
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
				"command":      {"/slack-slash-command"},
				"text":         {"invite-restricted user@example.com Tom Smith"},
				"user_name":    {"requesting_user"},
			}

			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			fakeSlackAPI.InviteRestrictedReturns(errors.New("failed to invite user"))

			w := httptest.NewRecorder()
			c = config.NewLocalConfig(
				"fake-slack-auth-token",
				"/slack-slash-command",
				"slack-team-name",
				"slack-user-id",
				"audit-log-channel-id",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
			)
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
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

	Describe("help", func() {
		It("responds to Slack with the help text", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"help"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(w.Body.String()).ShouldNot(BeEmpty())
		})
	})

	Describe("info", func() {
		It("asks Slack for the list of users", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"info user@example.com"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.GetUsersCallCount()).Should(Equal(1))
		})

		It("responds to Slack with a message about an unknown user", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"info user@example.com"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			fakeSlackAPI.GetUsersReturns([]slack.User{}, nil)
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(w.Body.String()).Should(Equal("There is no user here with the email address 'user@example.com'. You can invite them to Slack as a guest or a restricted account. Type `/slack-slash-command help` for more information."))
		})

		It("responds to Slack with a message about an unknown user with an uninvitable domain", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"info user@uninvitable-domain.com"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			fakeSlackAPI.GetUsersReturns([]slack.User{}, nil)
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(w.Body.String()).Should(Equal("There is no user here with the email address 'user@uninvitable-domain.com'. uninvitable-domain-message"))
		})

		It("responds to Slack with a message about a full member", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"info user@example.com"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					Name: "tsmith",
					Profile: slack.UserProfile{
						Email:     "user@example.com",
						FirstName: "Tom",
						LastName:  "Smith",
					},
					IsRestricted:      false,
					IsUltraRestricted: false,
				},
			}, nil)
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(w.Body.String()).Should(Equal("Tom Smith (user@example.com) is a Slack full member, with the username <@tsmith>."))
		})

		It("responds to Slack with a message about a restricted account", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"info user@example.com"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					Name: "tsmith",
					Profile: slack.UserProfile{
						Email:     "user@example.com",
						FirstName: "Tom",
						LastName:  "Smith",
					},
					IsRestricted:      true,
					IsUltraRestricted: false,
				},
			}, nil)
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(w.Body.String()).Should(Equal("Tom Smith (user@example.com) is a Slack restricted account, with the username <@tsmith>."))
		})

		It("responds to Slack with a message about a single-channel guest", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"info user@example.com"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					Name: "tsmith",
					Profile: slack.UserProfile{
						Email:     "user@example.com",
						FirstName: "Tom",
						LastName:  "Smith",
					},
					IsRestricted:      false,
					IsUltraRestricted: true,
				},
			}, nil)
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(w.Body.String()).Should(Equal("Tom Smith (user@example.com) is a Slack single-channel guest, with the username <@tsmith>."))
		})

		It("responds to Slack when it can't get the list of users", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"info user@example.com"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			fakeSlackAPI.GetUsersReturns([]slack.User{}, errors.New("network error"))
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(w.Body.String()).Should(Equal("Failed to look up user@example.com: network error"))
		})

		It("posts a message to the configured audit log channel on success", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"info user@example.com"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					Name: "tsmith",
					Profile: slack.UserProfile{
						Email:     "user@example.com",
						FirstName: "Tom",
						LastName:  "Smith",
					},
					IsRestricted:      true,
					IsUltraRestricted: true,
				},
			}, nil)
			c = config.NewLocalConfig(
				"fake-slack-auth-token",
				"/slack-slash-command",
				"slack-team-name",
				"slack-user-id",
				"audit-log-channel-id",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
			)
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(1))

			expectedParams := slack.NewPostMessageParameters()
			expectedParams.AsUser = true
			expectedParams.Parse = "full"

			actualChannelID, actualText, actualParams := fakeSlackAPI.PostMessageArgsForCall(0)
			Ω(actualChannelID).Should(Equal("audit-log-channel-id"))
			Ω(actualText).Should(Equal("@requesting_user requested info on 'user@example.com' at 2014-01-31 10:59:53 +0000 UTC, which was successful."))
			Ω(actualParams).Should(Equal(expectedParams))
		})

		It("posts a message to the configured audit log channel on failure", func() {
			v := url.Values{
				"token":        {"some-token"},
				"channel_id":   {"C1234567890"},
				"channel_name": {"channel-name"},
				"command":      {"/slack-slash-command"},
				"text":         {"info user@example.com"},
				"user_name":    {"requesting_user"},
			}
			reqBody := strings.NewReader(v.Encode())
			r, err := http.NewRequest("POST", "http://localhost", reqBody)
			Ω(err).ShouldNot(HaveOccurred())

			r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

			w := httptest.NewRecorder()
			fakeSlackAPI := &slackapifakes.FakeSlackAPI{}
			fakeSlackAPI.GetUsersReturns([]slack.User{}, errors.New("network error"))
			c = config.NewLocalConfig(
				"fake-slack-auth-token",
				"/slack-slash-command",
				"slack-team-name",
				"slack-user-id",
				"audit-log-channel-id",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
			)
			h := handler.New(c, fakeSlackAPI, fakeClock, lager.NewLogger("fakelogger"))
			h.ServeHTTP(w, r)

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(1))

			expectedParams := slack.NewPostMessageParameters()
			expectedParams.AsUser = true
			expectedParams.Parse = "full"

			actualChannelID, actualText, actualParams := fakeSlackAPI.PostMessageArgsForCall(0)
			Ω(actualChannelID).Should(Equal("audit-log-channel-id"))
			Ω(actualText).Should(Equal("@requesting_user requested info on 'user@example.com' at 2014-01-31 10:59:53 +0000 UTC, which failed with error: network error"))
			Ω(actualParams).Should(Equal(expectedParams))
		})
	})
})

func newGroup(name, id string) slack.Group {
	group := slack.Group{}
	group.Name = name
	group.ID = id

	return group
}
