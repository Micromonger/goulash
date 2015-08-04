package action_test

import (
	"github.com/pivotalservices/goulash/action"
	"github.com/pivotalservices/goulash/slackapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Action", func() {
	Describe("New", func() {
		var a action.Action

		It("supports creating an info action", func() {
			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"info user@example.com",
			)

			Ω(a).To(Equal(action.NewInfo([]string{"user@example.com"}, "commander-name")))
		})

		It("supports creating an invite-guest action", func() {
			expectedChannel := slackapi.NewChannel("channel-name", "channel-id")
			a = action.New(
				expectedChannel,
				"commander-name",
				"commander-id",
				"invite-guest user@example.com Tom Smith",
			)

			Ω(a).To(Equal(action.NewInvite(
				[]string{"user@example.com", "Tom", "Smith"},
				"invite-guest",
				expectedChannel,
				"commander-name",
			)))
		})

		It("supports creating an invite-restricted action", func() {
			expectedChannel := slackapi.NewChannel("channel-name", "channel-id")
			a = action.New(
				expectedChannel,
				"commander-name",
				"commander-id",
				"invite-restricted user@example.com Tom Smith",
			)

			Ω(a).To(Equal(action.NewInvite(
				[]string{"user@example.com", "Tom", "Smith"},
				"invite-restricted",
				expectedChannel,
				"commander-name",
			)))
		})

		It("supports creating a disable user action", func() {
			a = action.New(
				slackapi.NewChannel("channel-name", "channel-id"),
				"commander-name",
				"commander-id",
				"disable-user user@example.com",
			)

			Ω(a).To(Equal(action.NewDisableUser([]string{"user@example.com"}, "commander-name")))
		})
	})
})
