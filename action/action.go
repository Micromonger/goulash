package action

import (
	"strings"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
)

// Action represents an action that is able to be performed by the server.
type Action interface {
	Do() (string, error)
}

// GuardedAction is an Action with prerequisites described in Check().
type GuardedAction interface {
	Check() error
}

// AuditableAction is an Action that should have an audit log entry created.
type AuditableAction interface {
	AuditMessage() string
}

// New creates a new Action based on the command provided.
func New(
	channel slackapi.Channel,
	commanderName string,
	commanderID string,
	text string,

	config config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) Action {
	command, commandParams := commandAndParams(text)

	switch command {
	case "help":
		return help{
			config: config,
		}

	case "info":
		return userInfo{
			params:         commandParams,
			requestingUser: commanderName,

			config: config,
			api:    api,
			logger: logger,
		}

	case "invite-guest":
		return inviteGuest{
			params:       commandParams,
			channel:      channel,
			invitingUser: commanderName,

			config: config,
			api:    api,
			logger: logger,
		}

	case "invite-restricted":
		return inviteRestricted{
			params:       commandParams,
			channel:      channel,
			invitingUser: commanderName,

			config: config,
			api:    api,
			logger: logger,
		}

	default:
		return help{}
	}
}

func commandAndParams(text string) (string, []string) {
	var command string
	var commandParams []string

	if commandSep := strings.IndexByte(text, 0x20); commandSep > 0 {
		command = text[:commandSep]
		commandParams = strings.Split(text[commandSep+1:], " ")
	} else {
		command = text
	}

	return command, commandParams
}

func uninvitableEmail(emailAddress string, uninvitableDomain string) bool {
	return len(uninvitableDomain) > 0 && strings.HasSuffix(emailAddress, uninvitableDomain)
}
