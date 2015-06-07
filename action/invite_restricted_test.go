package action_test

import (
	"errors"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/action"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"

	fakeslackapi "github.com/pivotalservices/goulash/slackapi/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InviteRestricted", func() {
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

	Describe("Check", func() {
		It("returns nil", func() {
			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"invite-restricted user@example.com Tom Smith",
			)
			ga := a.(action.GuardedAction)
			Ω(ga.Check(c, nil, logger)).To(BeNil())
		})

		It("returns an error when the email has an uninvitable domain", func() {
			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"invite-restricted user@uninvitable-domain.com Tom Smith",
			)
			ga := a.(action.GuardedAction)
			err := ga.Check(c, nil, logger)
			Ω(err).To(HaveOccurred())
		})

		It("returns an error when the channel is not visible", func() {
			fakeChannel := &fakeslackapi.FakeChannel{}
			fakeChannel.VisibleReturns(false)

			a = action.New(
				fakeChannel,
				"commander-name",
				"commander-id",
				"invite-restricted user@example.com Tom Smith",
			)
			ga := a.(action.GuardedAction)
			Ω(ga.Check(c, nil, logger)).To(BeAssignableToTypeOf(action.ChannelNotVisibleErr{}))
		})

		It("returns an error when the email address is missing", func() {
			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"invite-restricted",
			)
			ga := a.(action.GuardedAction)
			err := ga.Check(c, nil, logger)
			Ω(err).To(BeAssignableToTypeOf(action.NewMissingEmailParameterErr("/slack-slash-command")))
		})
	})

	Describe("Do", func() {
		It("attempts to invite a restricted account", func() {
			fakeSlackAPI := &fakeslackapi.FakeSlackAPI{}

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"invite-restricted user@example.com Tom Smith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).NotTo(HaveOccurred())
			Ω(result).To(Equal("@commander-name invited Tom Smith (user@example.com) as a restricted account to 'channel-name'"))

			Ω(fakeSlackAPI.InviteRestrictedCallCount()).Should(Equal(1))

			actualTeamName, actualChannelID, actualFirstName, actualLastName, actualEmailAddress := fakeSlackAPI.InviteRestrictedArgsForCall(0)
			Ω(actualTeamName).Should(Equal("slack-team-name"))
			Ω(actualChannelID).Should(Equal("channel-id"))
			Ω(actualFirstName).Should(Equal("Tom"))
			Ω(actualLastName).Should(Equal("Smith"))
			Ω(actualEmailAddress).Should(Equal("user@example.com"))
		})

		It("returns an error on failure", func() {
			fakeSlackAPI := &fakeslackapi.FakeSlackAPI{}
			fakeSlackAPI.InviteRestrictedReturns(errors.New("failed"))

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"invite-restricted user@example.com Tom Smith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(result).To(Equal("Failed to invite Tom Smith (user@example.com) as a restricted account to 'channel-name': failed"))
			Ω(err).To(HaveOccurred())
			Ω(err.Error()).To(Equal("failed"))
		})

		It("returns nil on success", func() {
			fakeSlackAPI := &fakeslackapi.FakeSlackAPI{}

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"invite-restricted user@example.com Tom Smith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).NotTo(HaveOccurred())
			Ω(result).To(Equal("@commander-name invited Tom Smith (user@example.com) as a restricted account to 'channel-name'"))
		})
	})
})
