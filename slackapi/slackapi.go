package slackapi

import "github.com/pivotalservices/slack"

const (
	// PrivateGroupName holds the name Slack provides for a Slash Command sent
	// from a group which is private.
	PrivateGroupName = "privategroup"
)

//go:generate counterfeiter . SlackAPI

// SlackAPI defines the set of methods we expect to call on slack.Slack. This
// allows us to fake it for testing purposes.
type SlackAPI interface {
	// channel
	PostMessage(channelID string, text string, params slack.PostMessageParameters) (channel string, timestamp string, err error)

	// admin
	InviteGuest(teamname string, channelID string, firstName string, lastName string, emailAddress string) error
	InviteRestricted(teamname, channelID, firstName, lastName, emailAddress string) error

	// groups
	GetGroups(excludeArchived bool) ([]slack.Group, error)

	// im
	OpenIMChannel(userID string) (bool, bool, string, error)

	// users
	DisableUser(teamName string, user string) error
	GetUsers() ([]slack.User, error)
}
