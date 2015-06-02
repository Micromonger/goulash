package action_test

import (
	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/action"
	"github.com/pivotalservices/goulash/slackapi"

	fakeslackapi "github.com/pivotalservices/goulash/slackapi/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InviteGuest", func() {
	Describe("Check", func() {
		var (
			a            action.Action
			fakeSlackAPI *fakeslackapi.FakeSlackAPI
			logger       lager.Logger
		)

		BeforeEach(func() {
			fakeSlackAPI = &fakeslackapi.FakeSlackAPI{}
			logger = lager.NewLogger("testlogger")
		})

		It("returns nil", func() {
			a = action.New(
				slackapi.NewChannel("channel-id", "channel-name"),
				"commander-name",
				"commander-id",
				"invite-guest user@example.com",
				fakeSlackAPI,
				"slack-team-name",
				"slack-user-id",
				"/slack-slash-command",
				"uninvitable-domain",
				"uninvitable-message",
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
				fakeSlackAPI,
				"slack-team-name",
				"slack-user-id",
				"/slack-slash-command",
				"uninvitable-domain.com",
				"uninvitable-message",
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
				fakeSlackAPI,
				"slack-team-name",
				"slack-user-id",
				"/slack-slash-command",
				"uninvitable-domain.com",
				"uninvitable-message",
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
				fakeSlackAPI,
				"slack-team-name",
				"slack-user-id",
				"/slack-slash-command",
				"uninvitable-domain.com",
				"uninvitable-message",
				logger,
			)
			ga := a.(action.GuardedAction)
			err := ga.Check()
			Ω(err).Should(BeAssignableToTypeOf(action.NewMissingEmailParameterErr("/slack-slash-command")))
		})
	})
})
