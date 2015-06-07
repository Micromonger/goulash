package action_test

import (
	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/action"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"

	fakeslackapi "github.com/pivotalservices/goulash/slackapi/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InviteGuest", func() {
	Describe("Check", func() {
		var (
			a            action.Action
			c            config.Config
			fakeSlackAPI *fakeslackapi.FakeSlackAPI
			logger       lager.Logger
		)

		BeforeEach(func() {
			fakeSlackAPI = &fakeslackapi.FakeSlackAPI{}
			c = config.NewLocalConfig(
				"slack-auth-token",
				"/slack-slash-command",
				"slack-team-name",
				"slack-user-id",
				"audit-log-channel-id",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
			)

			logger = lager.NewLogger("testlogger")
		})

		It("returns nil", func() {
			a = action.New(
				slackapi.NewChannel("channel-id", "channel-name"),
				"commander-name",
				"commander-id",
				"invite-guest user@example.com",
				c,
				fakeSlackAPI,
				logger,
			)
			ga := a.(action.GuardedAction)
			Ω(ga.Check()).To(BeNil())
		})

		It("returns an error when the email has an uninvitable domain", func() {
			a = action.New(
				slackapi.NewChannel("channel-id", "channel-name"),
				"commander-name",
				"commander-id",
				"invite-guest user@uninvitable-domain.com",
				c,
				fakeSlackAPI,
				logger,
			)
			ga := a.(action.GuardedAction)
			err := ga.Check()
			Ω(err).To(HaveOccurred())
		})

		It("returns an error when the channel is not visible", func() {
			fakeChannel := &fakeslackapi.FakeChannel{}
			fakeChannel.VisibleReturns(false)

			a = action.New(
				fakeChannel,
				"commander-name",
				"commander-id",
				"invite-guest user@example.com",
				c,
				fakeSlackAPI,
				logger,
			)
			ga := a.(action.GuardedAction)
			Ω(ga.Check()).To(BeAssignableToTypeOf(action.ChannelNotVisibleErr{}))
		})

		It("returns an error when the email address is missing", func() {
			a = action.New(
				slackapi.NewChannel("channel-id", "channel-name"),
				"commander-name",
				"commander-id",
				"invite-guest",
				c,
				fakeSlackAPI,
				logger,
			)
			ga := a.(action.GuardedAction)
			err := ga.Check()
			Ω(err).To(BeAssignableToTypeOf(action.NewMissingEmailParameterErr("/slack-slash-command")))
		})
	})
})
