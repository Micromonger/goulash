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

var _ = Describe("DisableUser", func() {
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
		It("attempts to disable the user if they can be found by name", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID:   "U1234",
					Name: "tsmith",
				},
			}, nil)

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user @tsmith",
			)

			_, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).NotTo(HaveOccurred())

			Ω(fakeSlackAPI.GetUsersCallCount()).Should(Equal(1))
			Ω(fakeSlackAPI.DisableUserCallCount()).Should(Equal(1))
		})

		It("attempts to disable the user if they can be found by email", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID: "U1234",
					Profile: slack.UserProfile{
						Email: "user@example.com",
					},
				},
			}, nil)

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			_, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).NotTo(HaveOccurred())

			Ω(fakeSlackAPI.GetUsersCallCount()).Should(Equal(1))
			Ω(fakeSlackAPI.DisableUserCallCount()).Should(Equal(1))
		})

		It("returns an error if the GetUsers call fails", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{}, errors.New("error"))

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).To(HaveOccurred())
			Ω(err.Error()).To(Equal("error"))
			Ω(result).To(Equal("Failed to disable user 'user@example.com': error"))
		})

		It("returns an error if the user cannot be found", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{}, nil)

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).To(HaveOccurred())
			Ω(err.Error()).To(Equal("Unable to find user matching 'user@example.com'."))
			Ω(result).To(Equal("Failed to disable user: Unable to find user matching 'user@example.com'."))
		})

		It("returns an error when disabling the user fails", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID: "U1234",
					Profile: slack.UserProfile{
						Email: "user@example.com",
					},
				},
			}, nil)

			fakeSlackAPI.DisableUserReturns(errors.New("failed"))

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(result).To(Equal("Failed to disable user 'user@example.com': failed"))
			Ω(err).To(HaveOccurred())
			Ω(err.Error()).To(Equal("failed"))
		})

		It("returns nil on success", func() {
			fakeSlackAPI.GetUsersReturns([]slack.User{
				{
					ID: "U1234",
					Profile: slack.UserProfile{
						Email: "user@example.com",
					},
				},
			}, nil)

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).NotTo(HaveOccurred())
			Ω(result).To(Equal("Successfully disabled user 'user@example.com'"))
		})

	})

	Describe("AuditMessage", func() {
		It("exists", func() {
			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			aa, ok := a.(action.AuditableAction)
			Ω(ok).Should(BeTrue())

			Ω(aa.AuditMessage(fakeSlackAPI)).Should(Equal("@commander-name disabled user user@example.com"))
		})
	})
})
