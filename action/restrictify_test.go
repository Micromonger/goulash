package action_test

import (
	"errors"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/action"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
	"github.com/pivotalservices/slack"

	fakeslackapi "github.com/pivotalservices/goulash/slackapi/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Restrictify", func() {
	Describe("Do", func() {
		var (
			c            config.Config
			fakeSlackAPI *fakeslackapi.FakeSlackAPI
			logger       lager.Logger
		)

		BeforeEach(func() {
			fakeSlackAPI = &fakeslackapi.FakeSlackAPI{}
			logger = lager.NewLogger("testlogger")
			c = config.NewLocalConfig(
				"slack-auth-token",
				"/slack-slash-command",
				"slack-team-name",
				"slack-user-id",
				"",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
			)
		})

		It("returns an error if the user can't be found due to error", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{}, errors.New("error"))

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"restrictify user@example.com",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).To(HaveOccurred())
			Ω(err.Error()).To(Equal("error"))
			Ω(result).To(Equal("Failed to restrictify user 'user@example.com': error"))

			Ω(fakeSlackAPI.SetRestrictedCallCount()).To(Equal(0))
		})

		It("returns an error if the user cannot be found", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"restrictify user@example.com",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).To(HaveOccurred())
			Ω(err.Error()).To(Equal("Unable to find user matching 'user@example.com'."))
			Ω(result).To(Equal("Failed to restrictify user 'user@example.com': Unable to find user matching 'user@example.com'."))

			Ω(fakeSlackAPI.SetRestrictedCallCount()).To(Equal(0))
		})

		It("returns an error if the user is a full user", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:                "U1234",
					Name:              "tsmith",
					IsRestricted:      false,
					IsUltraRestricted: false,
				},
			}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"restrictify @tsmith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).To(HaveOccurred())
			Ω(err.Error()).To(Equal("Full users cannot be restrictified."))
			Ω(result).To(Equal("Failed to restrictify user '@tsmith': Full users cannot be restrictified."))

			Ω(fakeSlackAPI.SetRestrictedCallCount()).To(Equal(0))
		})

		It("returns an error if the user is already a restricted account", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:                "U1234",
					Name:              "tsmith",
					IsRestricted:      true,
					IsUltraRestricted: false,
				},
			}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"restrictify @tsmith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).To(HaveOccurred())
			Ω(err.Error()).To(Equal("User is already a restricted account."))
			Ω(result).To(Equal("Failed to restrictify user '@tsmith': User is already a restricted account."))

			Ω(fakeSlackAPI.SetRestrictedCallCount()).To(Equal(0))
		})

		It("returns an error if the request comes from a direct message", func() {
			a := action.New(
				slackapi.NewChannel(slackapi.DirectMessageGroupName, "channel-id"),
				"commander-name",
				"commander-id",
				"restrictify @tsmith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).To(HaveOccurred())
			Ω(err.Error()).To(Equal("Cannot restrictify from a direct message. Try again from a channel or group."))
			Ω(result).To(Equal("Failed to restrictify user '@tsmith': Cannot restrictify from a direct message. Try again from a channel or group."))

			Ω(fakeSlackAPI.SetRestrictedCallCount()).To(Equal(0))
		})

		It("attempts to restrictify the user if they can be found by name", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:                "U1234",
					Name:              "tsmith",
					IsUltraRestricted: true,
				},
			}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"restrictify @tsmith",
			)

			_, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).NotTo(HaveOccurred())

			Ω(fakeSlackAPI.SetRestrictedCallCount()).To(Equal(1))

			actualTeamName, actualUserID := fakeSlackAPI.SetRestrictedArgsForCall(0)
			Ω(actualTeamName).To(Equal("slack-team-name"))
			Ω(actualUserID).To(Equal("U1234"))
		})

		It("attempts to restrictify the user if they can be found by email", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:                "U1234",
					IsUltraRestricted: true,
					Profile: slack.UserProfile{
						Email: "user@example.com",
					},
				},
			}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"restrictify user@example.com",
			)

			_, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).NotTo(HaveOccurred())

			Ω(fakeSlackAPI.SetRestrictedCallCount()).To(Equal(1))

			actualTeamName, actualUserID := fakeSlackAPI.SetRestrictedArgsForCall(0)
			Ω(actualTeamName).To(Equal("slack-team-name"))
			Ω(actualUserID).To(Equal("U1234"))
		})

		It("returns an error when restrictifying fails", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:                "U1234",
					Name:              "tsmith",
					IsUltraRestricted: true,
				},
			}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"restrictify @tsmith",
			)

			fakeSlackAPI.SetRestrictedReturns(errors.New("failed"))

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).To(HaveOccurred())
			Ω(result).To(Equal("Failed to restrictify user '@tsmith': failed"))
		})

		It("returns nil when restrictifying succeeds", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:                "U1234",
					Name:              "tsmith",
					IsUltraRestricted: true,
				},
			}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"restrictify @tsmith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).NotTo(HaveOccurred())
			Ω(result).To(Equal("Successfully restrictified user @tsmith"))
		})
	})

	Describe("AuditMessage", func() {
		var (
			fakeSlackAPI *fakeslackapi.FakeSlackAPI
		)

		BeforeEach(func() {
			fakeSlackAPI = &fakeslackapi.FakeSlackAPI{}
		})

		It("exists", func() {
			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"restrictify user@example.com",
			)

			aa, ok := a.(action.AuditableAction)
			Ω(ok).Should(BeTrue())

			Ω(aa.AuditMessage(fakeSlackAPI)).Should(Equal("@commander-name restrictified user 'user@example.com'"))
		})
	})
})
