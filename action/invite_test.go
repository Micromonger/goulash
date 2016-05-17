package action_test

import (
	"errors"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/action"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
	"github.com/pivotalservices/goulash/slackapi/slackapifakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Invite", func() {
	var (
		a            action.Action
		c            config.Config
		fakeSlackAPI *slackapifakes.FakeSlackAPI
		logger       lager.Logger
	)

	BeforeEach(func() {
		fakeSlackAPI = &slackapifakes.FakeSlackAPI{}
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

	Describe("Do", func() {
		It("returns an error when the email has an uninvitable domain", func() {
			expectedErr := action.NewUninvitableDomainErr("uninvitable-domain.com", "uninvitable-domain-message", "/slack-slash-command")

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"invite-guest user@uninvitable-domain.com Tom Smith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(BeAssignableToTypeOf(expectedErr))
			Ω(result).Should(Equal(expectedErr.Error()))

			Ω(fakeSlackAPI.InviteGuestCallCount()).Should(Equal(0))
		})

		It("returns an error when the channel is not visible", func() {
			expectedErr := action.NewChannelNotVisibleErr("slack-user-id")

			fakeChannel := &slackapifakes.FakeChannel{}
			fakeChannel.VisibleReturns(false)

			a = action.New(
				fakeChannel,
				"commander-name",
				"commander-id",
				"invite-guest user@example.com Tom Smith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(BeAssignableToTypeOf(expectedErr))
			Ω(result).Should(Equal(expectedErr.Error()))

			Ω(fakeSlackAPI.InviteGuestCallCount()).Should(Equal(0))
		})

		It("returns an error when the email address is missing", func() {
			expectedErr := action.NewMissingEmailParameterErr("/slack-slash-command")

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"invite-guest",
			)

			result, err := a.Do(c, &slackapifakes.FakeSlackAPI{}, logger)
			Ω(err).Should(BeAssignableToTypeOf(expectedErr))
			Ω(result).Should(Equal(expectedErr.Error()))

			Ω(fakeSlackAPI.InviteGuestCallCount()).Should(Equal(0))
		})

		It("attempts to invite a single-channel guest", func() {
			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"invite-guest user@example.com Tom Smith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(result).Should(Equal("Successfully invited Tom Smith (user@example.com) as a single-channel guest to 'channel-name'"))

			Ω(fakeSlackAPI.InviteGuestCallCount()).Should(Equal(1))

			actualTeamName, actualChannelID, actualFirstName, actualLastName, actualEmailAddress := fakeSlackAPI.InviteGuestArgsForCall(0)
			Ω(actualTeamName).Should(Equal("slack-team-name"))
			Ω(actualChannelID).Should(Equal("channel-id"))
			Ω(actualFirstName).Should(Equal("Tom"))
			Ω(actualLastName).Should(Equal("Smith"))
			Ω(actualEmailAddress).Should(Equal("user@example.com"))
		})

		It("attempts to invite a restricted account", func() {
			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"invite-restricted user@example.com Tom Smith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(result).Should(Equal("Successfully invited Tom Smith (user@example.com) as a restricted account to 'channel-name'"))

			Ω(fakeSlackAPI.InviteRestrictedCallCount()).Should(Equal(1))

			actualTeamName, actualChannelID, actualFirstName, actualLastName, actualEmailAddress := fakeSlackAPI.InviteRestrictedArgsForCall(0)
			Ω(actualTeamName).Should(Equal("slack-team-name"))
			Ω(actualChannelID).Should(Equal("channel-id"))
			Ω(actualFirstName).Should(Equal("Tom"))
			Ω(actualLastName).Should(Equal("Smith"))
			Ω(actualEmailAddress).Should(Equal("user@example.com"))
		})

		It("returns alternate success on 'already_invited' error", func() {
			fakeSlackAPI.InviteGuestReturns(errors.New("failed: already_invited"))

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"invite-guest user@example.com Tom Smith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(result).Should(Equal("Successfully invited Tom Smith (user@example.com) as a single-channel guest to 'channel-name'"))
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns an error on failure", func() {
			fakeSlackAPI.InviteGuestReturns(errors.New("failed"))

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"invite-guest user@example.com Tom Smith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(result).Should(Equal("Failed to invite Tom Smith (user@example.com) as a single-channel guest to 'channel-name': failed"))
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("failed"))
		})

		It("returns nil on success", func() {
			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"invite-guest user@example.com Tom Smith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(result).Should(Equal("Successfully invited Tom Smith (user@example.com) as a single-channel guest to 'channel-name'"))
		})

		It("attempts to invite a single-channel guest when the args are padded with extra spaces", func() {
			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"invite-guest user@example.com  Tom  Smith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(result).Should(Equal("Successfully invited Tom Smith (user@example.com) as a single-channel guest to 'channel-name'"))

			Ω(fakeSlackAPI.InviteGuestCallCount()).Should(Equal(1))

			actualTeamName, actualChannelID, actualFirstName, actualLastName, actualEmailAddress := fakeSlackAPI.InviteGuestArgsForCall(0)
			Ω(actualTeamName).Should(Equal("slack-team-name"))
			Ω(actualChannelID).Should(Equal("channel-id"))
			Ω(actualFirstName).Should(Equal("Tom"))
			Ω(actualLastName).Should(Equal("Smith"))
			Ω(actualEmailAddress).Should(Equal("user@example.com"))
		})
	})
})
