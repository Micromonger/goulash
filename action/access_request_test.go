package action_test

import (
	"errors"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/action"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
	"github.com/pivotalservices/goulash/slackapi/slackapifakes"
	"github.com/pivotalservices/slack"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AccessRequest", func() {
	var (
		c            config.Config
		fakeSlackAPI *slackapifakes.FakeSlackAPI
		logger       lager.Logger
	)

	BeforeEach(func() {
		fakeSlackAPI = &slackapifakes.FakeSlackAPI{}
	})

	Describe("Do", func() {
		BeforeEach(func() {
			logger = lager.NewLogger("testlogger")
			c = config.NewLocalConfig(
				"slack-auth-token",
				"/slack-slash-command",
				"slack-team-name",
				"slack-user-id",
				"audit-log-channel-id",
				"uninvitable-domain.com",
				"uninvitable-domain-message",
			)
		})

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
				"request-access #channel-name",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("Sorry, you don't have access to that function."))
			Ω(result).Should(Equal("Failed to request access to #channel-name: Sorry, you don't have access to that function."))

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
				"request-access #channel-name",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("Sorry, you don't have access to that function."))
			Ω(result).Should(Equal("Failed to request access to #channel-name: Sorry, you don't have access to that function."))

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(0))
		})

		It("returns an error if the GetUserInfo call fails", func() {
			fakeSlackAPI.GetUserInfoReturns(&slack.User{}, errors.New("get-user-info-err"))

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"request-access #channel-name",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("get-user-info-err"))
			Ω(result).Should(Equal("Failed to request access to #channel-name: get-user-info-err"))

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(0))
		})

		It("returns an error when getting channels returns an error", func() {
			fakeSlackAPI.GetUserInfoReturns(&slack.User{}, nil)
			fakeSlackAPI.GetChannelsReturns([]slack.Channel{}, errors.New("get-channels-err"))

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"request-access #channel-name",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("get-channels-err"))
			Ω(result).Should(Equal("Failed to request access to #channel-name: get-channels-err"))

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(0))
		})

		It("returns an error when the channel can't be found", func() {
			fakeSlackAPI.GetUserInfoReturns(&slack.User{}, nil)
			fakeSlackAPI.GetChannelsReturns([]slack.Channel{}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"request-access #channel-name",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("Channel '#channel-name' not found."))
			Ω(result).Should(Equal("Failed to request access to #channel-name: Channel '#channel-name' not found."))

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(0))
		})

		It("attempts to send a message to the channel for which access was requested when the channel can be found", func() {
			fakeSlackAPI.GetUserInfoReturns(&slack.User{}, nil)
			expectedChannel := slack.Channel{}
			expectedChannel.Name = "channel-name"
			expectedChannel.ID = "channel-id"
			fakeSlackAPI.GetChannelsReturns([]slack.Channel{expectedChannel}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"request-access #channel-name",
			)

			_, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(fakeSlackAPI.PostMessageCallCount()).Should(Equal(1))

			actualChannelID, actualText, actualParams := fakeSlackAPI.PostMessageArgsForCall(0)
			Ω(actualChannelID).Should(Equal("channel-id"))
			Ω(actualText).Should(Equal("@commander-name would like to be invited to this channel. To invite them, use `/invite @commander-name`"))
			Ω(actualParams.AsUser).Should(BeTrue())
			Ω(actualParams.Parse).Should(Equal("full"))
		})

		It("returns a positive result and nil on success", func() {
			fakeSlackAPI.GetUserInfoReturns(&slack.User{}, nil)
			expectedChannel := slack.Channel{}
			expectedChannel.Name = "channel-name"
			expectedChannel.ID = "channel-id"
			fakeSlackAPI.GetChannelsReturns([]slack.Channel{expectedChannel}, nil)

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"request-access #channel-name",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(result).Should(Equal("Successfully requested access to <#channel-name>."))
		})

		It("returns an error if the PostMessage call fails", func() {
			fakeSlackAPI.GetUserInfoReturns(&slack.User{}, nil)
			expectedChannel := slack.Channel{}
			expectedChannel.Name = "channel-name"
			expectedChannel.ID = "channel-id"
			fakeSlackAPI.GetChannelsReturns([]slack.Channel{expectedChannel}, nil)

			fakeSlackAPI.PostMessageReturns("channel", "timestamp", errors.New("post-message-err"))

			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"request-access #channel-name",
			)

			result, err := a.Do(c, fakeSlackAPI, logger)
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("post-message-err"))
			Ω(result).Should(Equal("Failed to request access to #channel-name: post-message-err"))
		})
	})

	Describe("AuditMessage", func() {
		It("exists", func() {
			a := action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"request-access #channel-name",
			)

			aa, ok := a.(action.AuditableAction)
			Ω(ok).Should(BeTrue())

			Ω(aa.AuditMessage(fakeSlackAPI)).Should(Equal("@commander-name requested access to #channel-name"))
		})
	})
})
