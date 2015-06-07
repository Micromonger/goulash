package action_test

import (
	"errors"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/slack"

	"github.com/pivotalservices/goulash/action"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"

	fakeslackapi "github.com/pivotalservices/goulash/slackapi/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserInfo", func() {
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

		It("asks Slack for the list of users", func() {
			a := action.New(
				slackapi.NewChannel("channel-id", "channel-name"),
				"commander-name",
				"commander-id",
				"info user@example.com",
			)

			a.Do(c, fakeSlackAPI, logger)

			Ω(fakeSlackAPI.GetUsersCallCount()).Should(Equal(1))
		})

		It("returns an error when it can't get the list of users from Slack", func() {
			a := action.New(
				slackapi.NewChannel("channel-id", "channel-name"),
				"commander-name",
				"commander-id",
				"info user@example.com",
			)

			fakeSlackAPI.GetUsersReturns([]slack.User{}, errors.New("network error"))

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(result).Should(Equal("Failed to look up user@example.com: network error"))
		})

		It("returns an error when no email address was given", func() {
			expectedErr := action.NewMissingEmailParameterErr("/slack-slash-command")

			a := action.New(
				slackapi.NewChannel("channel-id", "channel-name"),
				"commander-name",
				"commander-id",
				"info",
			)

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

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err).Should(BeAssignableToTypeOf(expectedErr))
			Ω(result).Should(Equal(expectedErr.Error()))
		})

		It("returns a result for an unknown user", func() {
			a := action.New(
				slackapi.NewChannel("channel-id", "channel-name"),
				"commander-name",
				"commander-id",
				"info user@example.com",
			)

			fakeSlackAPI.GetUsersReturns([]slack.User{}, nil)
			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(result).Should(Equal("There is no user here with the email address 'user@example.com'. You can invite them to Slack as a guest or a restricted account. Type `/slack-slash-command help` for more information."))
		})

		It("returns a result for a user with an uninvitable domain", func() {
			a := action.New(
				slackapi.NewChannel("channel-id", "channel-name"),
				"commander-name",
				"commander-id",
				"info user@uninvitable-domain.com",
			)

			fakeSlackAPI.GetUsersReturns([]slack.User{}, nil)
			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(result).Should(Equal("There is no user here with the email address 'user@uninvitable-domain.com'. uninvitable-domain-message"))
		})

		It("returns a result for a Slack 'full' member", func() {
			a := action.New(
				slackapi.NewChannel("channel-id", "channel-name"),
				"commander-name",
				"commander-id",
				"info user@example.com",
			)

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

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(result).Should(Equal("Tom Smith (user@example.com) is a Slack full member, with the username <@tsmith>."))
		})

		It("responds to Slack with a message about a restricted account", func() {
			a := action.New(
				slackapi.NewChannel("channel-id", "channel-name"),
				"commander-name",
				"commander-id",
				"info user@example.com",
			)

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

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(result).Should(Equal("Tom Smith (user@example.com) is a Slack restricted account, with the username <@tsmith>."))
		})

		It("responds to Slack with a message about a single-channel guest", func() {
			a := action.New(
				slackapi.NewChannel("channel-id", "channel-name"),
				"commander-name",
				"commander-id",
				"info user@example.com",
			)

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

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(result).Should(Equal("Tom Smith (user@example.com) is a Slack single-channel guest, with the username <@tsmith>."))
		})
	})
})
