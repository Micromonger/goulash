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

var _ = Describe("InviteGuest", func() {
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
			Ω(err).To(BeAssignableToTypeOf(expectedErr))
			Ω(result).To(Equal(expectedErr.Error()))

			Ω(fakeSlackAPI.InviteGuestCallCount()).Should(Equal(0))
		})

		It("returns an error when the channel is not visible", func() {
			expectedErr := action.NewChannelNotVisibleErr("slack-user-id")

			fakeChannel := &fakeslackapi.FakeChannel{}
			fakeChannel.VisibleReturns(false)

			a = action.New(
				fakeChannel,
				"commander-name",
				"commander-id",
				"invite-guest user@example.com Tom Smith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).To(BeAssignableToTypeOf(expectedErr))
			Ω(result).To(Equal(expectedErr.Error()))

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

			result, err := a.Do(c, &fakeslackapi.FakeSlackAPI{}, logger)
			Ω(err).To(BeAssignableToTypeOf(expectedErr))
			Ω(result).To(Equal(expectedErr.Error()))

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
			Ω(err).NotTo(HaveOccurred())
			Ω(result).To(Equal("@commander-name invited Tom Smith (user@example.com) as a guest to 'channel-name'"))

			Ω(fakeSlackAPI.InviteGuestCallCount()).Should(Equal(1))

			actualTeamName, actualChannelID, actualFirstName, actualLastName, actualEmailAddress := fakeSlackAPI.InviteGuestArgsForCall(0)
			Ω(actualTeamName).Should(Equal("slack-team-name"))
			Ω(actualChannelID).Should(Equal("channel-id"))
			Ω(actualFirstName).Should(Equal("Tom"))
			Ω(actualLastName).Should(Equal("Smith"))
			Ω(actualEmailAddress).Should(Equal("user@example.com"))
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
			Ω(result).To(Equal("Failed to invite Tom Smith (user@example.com) as a guest to 'channel-name': failed"))
			Ω(err).To(HaveOccurred())
			Ω(err.Error()).To(Equal("failed"))
		})

		It("returns nil on success", func() {
			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"invite-guest user@example.com Tom Smith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).NotTo(HaveOccurred())
			Ω(result).To(Equal("@commander-name invited Tom Smith (user@example.com) as a guest to 'channel-name'"))
		})
	})
})
