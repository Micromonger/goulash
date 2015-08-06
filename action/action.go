package action

import (
	"fmt"
	"strings"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
	"github.com/pivotalservices/slack"
)

// Action represents an action that is able to be performed by the server.
type Action interface {
	Do(config.Config, slackapi.SlackAPI, lager.Logger) (string, error)
}

// AuditableAction is an Action that should have an audit log entry created.
type AuditableAction interface {
	AuditMessage(slackapi.SlackAPI) string
}

// New creates a new Action based on the command provided.
func New(
	channel slackapi.Channel,
	commanderName string,
	commanderID string,
	text string,
) Action {
	commandAndParams := strings.Split(text, " ")
	command := commandAndParams[0]
	params := commandAndParams[1:]

	switch command {
	case "info":
		return NewInfo(params, commanderName)

	case "invite-guest", "invite-restricted":
		return NewInvite(params, command, channel, commanderName)

	case "disable-user":
		return NewDisableUser(params, commanderName)

	case "guestify":
		return NewGuestify(params, channel, commanderName)

	default:
		return help{}
	}
}

func uninvitableEmail(emailAddress string, uninvitableDomain string) bool {
	return len(uninvitableDomain) > 0 && strings.HasSuffix(emailAddress, uninvitableDomain)
}

func findUser(searchVal string, api slackapi.SlackAPI) (slack.User, error) {
	users, err := api.GetUsers()
	if err != nil {
		return slack.User{}, err
	}

	var foundUser slack.User
	for _, user := range users {
		if matches(searchVal, user.Profile.Email, fmt.Sprintf("@%s", user.Name)) {
			foundUser = user
			break
		}
	}

	if (foundUser == slack.User{}) {
		err = NewUserNotFoundErr(searchVal)
		return slack.User{}, err
	}

	return foundUser, nil
}

func matches(searchVal string, candidates ...string) bool {
	var match bool
	for _, candidate := range candidates {
		if searchVal == candidate {
			match = true
			break
		}
	}
	return match
}
