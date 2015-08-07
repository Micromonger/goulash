package slackapi

import "github.com/pivotalservices/slack"

const (
	// PrivateGroupName holds the name Slack provides for a Slash Command sent
	// from a group which is private.
	PrivateGroupName = "privategroup"

	// DirectMessageGroupName holds the name Slack provides for a Slash Command sent
	// from a direct message.
	DirectMessageGroupName = "directmessage"
)

//go:generate counterfeiter . SlackAPI

// SlackAPI defines the set of methods we expect to call on slack.Slack. This
// allows us to fake it for testing purposes.
type SlackAPI interface {
	// channel
	PostMessage(channelID string, text string, params slack.PostMessageParameters) (channel string, timestamp string, err error)
	GetChannels(excludeArchived bool) ([]slack.Channel, error)

	// admin
	InviteGuest(teamName string, channelID string, firstName string, lastName string, emailAddress string) error
	InviteRestricted(teamName, channelID, firstName, lastName, emailAddress string) error
	DisableUser(teamName string, user string) error
	SetUltraRestricted(teamName string, user string, channel string) error
	SetRestricted(teamName string, user string) error

	// groups
	GetGroups(excludeArchived bool) ([]slack.Group, error)

	// im
	OpenIMChannel(userID string) (bool, bool, string, error)

	// users
	GetUserInfo(userID string) (*slack.User, error)
	GetUsers() ([]slack.User, error)
}
