package slackapi_test

import (
	"github.com/pivotalservices/slack"

	"github.com/pivotalservices/goulash/slackapi"

	fakeslackapi "github.com/pivotalservices/goulash/slackapi/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Channel", func() {
	Describe("Name", func() {
		var fakeSlackAPI *fakeslackapi.FakeSlackAPI
		var channel slackapi.Channel

		BeforeEach(func() {
			fakeSlackAPI = &fakeslackapi.FakeSlackAPI{}
		})

		Describe("when the group's name is slackapi.PrivateGroupName", func() {
			BeforeEach(func() {
				channel = slackapi.NewChannel(slackapi.PrivateGroupName, "C1234")
			})

			It("tries to find the group's name in Slack, excluding archived groups", func() {
				channel.Name(fakeSlackAPI)
				Ω(fakeSlackAPI.GetGroupsCallCount()).Should(Equal(1))

				excludeArchived := fakeSlackAPI.GetGroupsArgsForCall(0)
				Ω(excludeArchived).Should(BeTrue())
			})

			It("only hits the Slack API once", func() {
				channel.Name(fakeSlackAPI)
				channel.Name(fakeSlackAPI)
				Ω(fakeSlackAPI.GetGroupsCallCount()).Should(Equal(1))
			})

			Describe("when the group is found in Slack", func() {
				BeforeEach(func() {
					group := slack.Group{}
					group.Name = "channel-name"
					group.ID = "C1234"
					fakeSlackAPI.GetGroupsReturns([]slack.Group{group}, nil)
				})

				It("returns the name associated with the found group", func() {
					Ω(channel.Name(fakeSlackAPI)).Should(Equal("channel-name"))
				})
			})

			Describe("when the group's name is not found in Slack", func() {
				BeforeEach(func() {
					group := slack.Group{}
					group.Name = "other-channel-name"
					group.ID = "C9999"
					fakeSlackAPI.GetGroupsReturns([]slack.Group{group}, nil)
				})

				It("returns slackapi.PrivateGroupName", func() {
					Ω(channel.Name(fakeSlackAPI)).Should(Equal(slackapi.PrivateGroupName))
				})
			})
		})

		Describe("when the group's name is not slackapi.PrivateGroupName", func() {
			BeforeEach(func() {
				channel = slackapi.NewChannel("channel-name", "C1234")
			})

			It("returns the name", func() {
				Ω(channel.Name(fakeSlackAPI)).Should(Equal("channel-name"))
			})

			It("does not call Slack to find the group's name", func() {
				channel.Name(fakeSlackAPI)
				Ω(fakeSlackAPI.GetGroupsCallCount()).Should(Equal(0))
			})
		})
	})
})
