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

var _ = Describe("Groups", func() {
	var (
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
		It("returns an error if the commanding user is a single-channel guest", func() {
			fakeSlackAPI.GetUserInfoReturns(&slack.User{
				ID:                "commander-id",
				IsRestricted:      true,
				IsUltraRestricted: true,
			}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"groups",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("Sorry, you don't have access to that function."))
			Ω(result).Should(Equal("Failed to list the groups I'm in: Sorry, you don't have access to that function."))

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(0))
		})

		It("returns an error if the commanding user is a restricted account", func() {
			fakeSlackAPI.GetUserInfoReturns(&slack.User{
				ID:                "commander-id",
				IsRestricted:      true,
				IsUltraRestricted: false,
			}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"groups",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("Sorry, you don't have access to that function."))
			Ω(result).Should(Equal("Failed to list the groups I'm in: Sorry, you don't have access to that function."))

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(0))
		})

		It("returns an error if the GetUserInfo call fails", func() {
			fakeSlackAPI.GetUserInfoReturns(&slack.User{}, errors.New("get-user-info-err"))

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"groups",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("get-user-info-err"))
			Ω(result).Should(Equal("Failed to list the groups I'm in: get-user-info-err"))

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(0))
		})

		It("attempts to get groups", func() {
			fakeSlackAPI.GetUserInfoReturns(&slack.User{}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"groups",
			)

			_, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(fakeSlackAPI.GetGroupsCallCount()).Should(Equal(1))
			actualExcludeArchived := fakeSlackAPI.GetGroupsArgsForCall(0)
			Ω(actualExcludeArchived).Should(BeTrue())
		})

		It("attempts to send a direct message with a sorted list of the returned groups", func() {
			fakeSlackAPI.GetUserInfoReturns(&slack.User{}, nil)
			fakeSlackAPI.GetGroupsReturns([]slack.Group{
				newGroup("group-2", "slack-user-id"),
				newGroup("group-1", "slack-user-id"),
				newGroup("group-3", "slack-user-id"),
			}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"groups",
			)

			_, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(1))

			actualChannelID, actualText, actualParams := fakeSlackAPI.PostMessageArgsForCall(0)
			Ω(actualChannelID).Should(Equal("commander-name"))
			Ω(actualText).Should(Equal("I'm in the following groups:\n\ngroup-1\ngroup-2\ngroup-3"))
			Ω(actualParams.AsUser).Should(BeTrue())
		})

		It("returns a positive result and nil on success", func() {
			fakeSlackAPI.GetUserInfoReturns(&slack.User{}, nil)
			fakeSlackAPI.GetGroupsReturns([]slack.Group{
				newGroup("group-1", "slack-user-id"),
			}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"groups",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(result).Should(Equal("Successfully sent a list of the groups I'm in as a direct message."))
		})

		It("returns an error if the PostMessage call fails", func() {
			fakeSlackAPI.GetUserInfoReturns(&slack.User{}, nil)
			fakeSlackAPI.GetGroupsReturns([]slack.Group{
				newGroup("group-2", "slack-user-id"),
				newGroup("group-1", "slack-user-id"),
				newGroup("group-3", "slack-user-id"),
			}, nil)

			fakeSlackAPI.PostMessageReturns("channel", "timestamp", errors.New("failed"))

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"groups",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("failed"))
			Ω(result).Should(Equal("Failed to list the groups I'm in: failed"))
		})

		It("returns an error if the GetGroups call fails", func() {
			fakeSlackAPI.GetUserInfoReturns(&slack.User{}, nil)
			fakeSlackAPI.GetGroupsReturns([]slack.Group{}, errors.New("failed"))

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"groups",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("failed"))
			Ω(result).Should(Equal("Failed to list the groups I'm in: failed"))
		})
	})

	Describe("AuditMessage", func() {
		It("exists", func() {
			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"groups",
			)

			aa, ok := a.(action.AuditableAction)
			Ω(ok).Should(BeTrue())

			Ω(aa.AuditMessage(fakeSlackAPI)).Should(Equal("@commander-name requested the groups that I'm in"))
		})
	})
})

func newGroup(name, member string) slack.Group {
	group := slack.Group{IsGroup: true}
	group.Name = name
	group.Members = []string{member}
	return group
}
