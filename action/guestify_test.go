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

var _ = Describe("Guestify", func() {
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
				"guestify user@example.com",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("error"))
			Ω(result).Should(Equal("Failed to guestify user 'user@example.com': error"))

			Ω(fakeSlackAPI.SetUltraRestrictedCallCount()).Should(Equal(0))
		})

		It("returns an error if the user cannot be found", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"guestify user@example.com",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("Unable to find user matching 'user@example.com'."))
			Ω(result).Should(Equal("Failed to guestify user 'user@example.com': Unable to find user matching 'user@example.com'."))

			Ω(fakeSlackAPI.SetUltraRestrictedCallCount()).Should(Equal(0))
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
				"guestify @tsmith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("Full users cannot be guestified."))
			Ω(result).Should(Equal("Failed to guestify user '@tsmith': Full users cannot be guestified."))

			Ω(fakeSlackAPI.SetUltraRestrictedCallCount()).Should(Equal(0))
		})

		It("returns an error if the user is already a single-channel guest", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:                "U1234",
					Name:              "tsmith",
					IsRestricted:      false,
					IsUltraRestricted: true,
				},
			}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"guestify @tsmith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("User is already a single-channel guest."))
			Ω(result).Should(Equal("Failed to guestify user '@tsmith': User is already a single-channel guest."))

			Ω(fakeSlackAPI.SetUltraRestrictedCallCount()).Should(Equal(0))
		})

		It("returns an error if the request comes from a direct message", func() {
			a := action.New(
				slackapi.NewChannel(slackapi.DirectMessageGroupName, "channel-id"),
				"commander-name",
				"commander-id",
				"guestify @tsmith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("Cannot guestify from a direct message. Try again from a channel or group."))
			Ω(result).Should(Equal("Failed to guestify user '@tsmith': Cannot guestify from a direct message. Try again from a channel or group."))

			Ω(fakeSlackAPI.SetUltraRestrictedCallCount()).Should(Equal(0))
		})

		It("attempts to guestify the user if they can be found by name", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:           "U1234",
					Name:         "tsmith",
					IsRestricted: true,
				},
			}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"guestify @tsmith",
			)

			_, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(fakeSlackAPI.SetUltraRestrictedCallCount()).Should(Equal(1))

			actualTeamName, actualUserID, actualChannel := fakeSlackAPI.SetUltraRestrictedArgsForCall(0)
			Ω(actualTeamName).Should(Equal("slack-team-name"))
			Ω(actualUserID).Should(Equal("U1234"))
			Ω(actualChannel).Should(Equal("channel-id"))
		})

		It("attempts to guestify the user if they can be found by email", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:           "U1234",
					IsRestricted: true,
					Profile: slack.UserProfile{
						Email: "user@example.com",
					},
				},
			}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"guestify user@example.com",
			)

			_, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(fakeSlackAPI.SetUltraRestrictedCallCount()).Should(Equal(1))

			actualTeamName, actualUserID, actualChannel := fakeSlackAPI.SetUltraRestrictedArgsForCall(0)
			Ω(actualTeamName).Should(Equal("slack-team-name"))
			Ω(actualUserID).Should(Equal("U1234"))
			Ω(actualChannel).Should(Equal("channel-id"))
		})

		It("returns an error when guestifying fails", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:           "U1234",
					Name:         "tsmith",
					IsRestricted: true,
				},
			}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"guestify @tsmith",
			)

			fakeSlackAPI.SetUltraRestrictedReturns(errors.New("failed"))

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(result).Should(Equal("Failed to guestify user '@tsmith': failed"))
		})

		It("returns nil when guestifying succeeds", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:           "U1234",
					Name:         "tsmith",
					IsRestricted: true,
				},
			}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"guestify @tsmith",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(result).Should(Equal("Successfully guestified user @tsmith"))
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
				"guestify user@example.com",
			)

			aa, ok := a.(action.AuditableAction)
			Ω(ok).Should(BeTrue())

			Ω(aa.AuditMessage(fakeSlackAPI)).Should(Equal("@commander-name guestified user 'user@example.com'"))
		})
	})
})
