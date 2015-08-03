package action

import (
	"strings"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
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

	default:
		return help{}
	}
}

func uninvitableEmail(emailAddress string, uninvitableDomain string) bool {
	return len(uninvitableDomain) > 0 && strings.HasSuffix(emailAddress, uninvitableDomain)
}
