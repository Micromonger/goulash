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
		It("attempts to disable a user when given an email address", func() {
			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			_, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).NotTo(HaveOccurred())

			Ω(fakeSlackAPI.DisableUserCallCount()).Should(Equal(1))

			actualTeamName, actualUser := fakeSlackAPI.DisableUserArgsForCall(0)
			Ω(actualTeamName).Should(Equal("slack-team-name"))
			Ω(actualUser).Should(Equal("user@example.com"))
		})

		It("returns nil on success", func() {
			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).NotTo(HaveOccurred())
			Ω(result).To(Equal("Successfully disabled user@example.com"))
		})

		It("returns an error on failure", func() {
			fakeSlackAPI.DisableUserReturns(errors.New("failed"))

			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(result).To(Equal("Failed to disable user@example.com: failed"))
			Ω(err).To(HaveOccurred())
			Ω(err.Error()).To(Equal("failed"))
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
