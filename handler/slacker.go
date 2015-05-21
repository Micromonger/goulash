package handler

import "github.com/krishicks/slack"

//go:generate counterfeiter . SlackAPI

// SlackAPI defines the set of methods we expect to call on slack.Slack. This
// allows us to fake it for testing purposes.
type SlackAPI interface {
	PostMessage(channelID string, text string, params slack.PostMessageParameters) (channel string, timestamp string, err error)
}
