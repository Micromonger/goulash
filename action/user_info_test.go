package action_test

import (
	"errors"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/slack"

	"github.com/pivotalservices/goulash/action"

	fakeslackapi "github.com/pivotalservices/goulash/slackapi/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserInfo", func() {
	Describe("Do", func() {
		var (
			fakeSlackAPI *fakeslackapi.FakeSlackAPI
			logger       lager.Logger
		)

		BeforeEach(func() {
			fakeSlackAPI = &fakeslackapi.FakeSlackAPI{}
			logger = lager.NewLogger("testlogger")
		})

		It("asks Slack for the list of users", func() {
			a := action.New(
				"channel-id",
				"channel-name",
				"commander-name",
				"commander-id",
				"info user@example.com",

				fakeSlackAPI,
				"requesting_user",
				"slack-team-name",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
				logger,
			)

			a.Do()

			Ω(fakeSlackAPI.GetUsersCallCount()).To(Equal(1))
		})

		It("returns an error when it can't get the list of users from Slack", func() {
			a := action.New(
				"channel-id",
				"channel-name",
				"commander-name",
				"commander-id",
				"info user@example.com",

				fakeSlackAPI,
				"requesting_user",
				"slack-team-name",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
				logger,
			)

			fakeSlackAPI.GetUsersReturns([]slack.User{}, errors.New("network error"))

			result, err := a.Do()
			Ω(err).To(HaveOccurred())
			Ω(result).Should(Equal("Failed to look up user@example.com: network error"))
		})

		It("returns a result for an unknown user", func() {
			a := action.New(
				"channel-id",
				"channel-name",
				"commander-name",
				"commander-id",
				"info user@example.com",

				fakeSlackAPI,
				"requesting_user",
				"slack-team-name",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
				logger,
			)

			fakeSlackAPI.GetUsersReturns([]slack.User{}, nil)
			result, err := a.Do()
			Ω(err).To(HaveOccurred())
			Ω(result).Should(Equal("There is no user here with the email address 'user@example.com'. You can invite them to Slack as a guest or a restricted account. Type `/butler help` for more information."))
		})

		It("returns a result for a user with an uninvitable domain", func() {
			a := action.New(
				"channel-id",
				"channel-name",
				"commander-name",
				"commander-id",
				"info user@uninvitable-domain.com",

				fakeSlackAPI,
				"requesting_user",
				"slack-team-name",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
				logger,
			)

			fakeSlackAPI.GetUsersReturns([]slack.User{}, nil)
			result, err := a.Do()
			Ω(err).To(HaveOccurred())
			Ω(result).Should(Equal("There is no user here with the email address 'user@uninvitable-domain.com'. uninvitable-domain-message"))
		})

		It("returns a result for a Slack 'full' member", func() {
			a := action.New(
				"channel-id",
				"channel-name",
				"commander-name",
				"commander-id",
				"info user@example.com",

				fakeSlackAPI,
				"requesting_user",
				"slack-team-name",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
				logger,
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

			result, err := a.Do()
			Ω(err).NotTo(HaveOccurred())
			Ω(result).Should(Equal("Tom Smith (user@example.com) is a Slack full member, with the username <@tsmith>."))
		})

		It("responds to Slack with a message about a restricted account", func() {
			a := action.New(
				"channel-id",
				"channel-name",
				"commander-name",
				"commander-id",
				"info user@example.com",

				fakeSlackAPI,
				"requesting_user",
				"slack-team-name",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
				logger,
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

			result, err := a.Do()
			Ω(err).NotTo(HaveOccurred())
			Ω(result).Should(Equal("Tom Smith (user@example.com) is a Slack restricted account, with the username <@tsmith>."))
		})

		It("responds to Slack with a message about a single-channel guest", func() {
			a := action.New(
				"channel-id",
				"channel-name",
				"commander-name",
				"commander-id",
				"info user@example.com",

				fakeSlackAPI,
				"requesting_user",
				"slack-team-name",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
				logger,
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

			result, err := a.Do()
			Ω(err).NotTo(HaveOccurred())
			Ω(result).Should(Equal("Tom Smith (user@example.com) is a Slack single-channel guest, with the username <@tsmith>."))
		})
	})
})
