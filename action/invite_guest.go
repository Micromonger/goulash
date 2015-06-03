package action

import (
	"fmt"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
)

type inviteGuest struct {
	params       []string
	channel      slackapi.Channel
	invitingUser string

	config config.Config
	api    slackapi.SlackAPI
	logger lager.Logger
}

func (i inviteGuest) emailAddress() string {
	if len(i.params) > 0 {
		return i.params[0]
	}

	return ""
}

func (i inviteGuest) firstName() string {
	if len(i.params) > 1 {
		return i.params[1]
	}

	return ""
}

func (i inviteGuest) lastName() string {
	if len(i.params) > 2 {
		return i.params[2]
	}

	return ""
}

func (i inviteGuest) Check() error {
	if i.emailAddress() == "" {
		return NewMissingEmailParameterErr(i.config.SlackSlashCommand())
	}

	if uninvitableEmail(i.emailAddress(), i.config.UninvitableDomain()) {
		return NewUninvitableDomainErr(
			i.config.UninvitableDomain(),
			i.config.UninvitableMessage(),
			i.config.SlackSlashCommand(),
		)
	}

	if !i.channel.Visible(i.api) {
		return NewChannelNotVisibleErr(i.config.SlackUserID())
	}

	return nil
}

func (i inviteGuest) Do() (string, error) {
	var result string

	err := i.api.InviteGuest(
		i.config.SlackTeamName(),
		i.channel.ID(),
		i.firstName(),
		i.lastName(),
		i.emailAddress(),
	)
	if err != nil {
		i.logger.Error("failed-inviting-single-channel-user", err)
		result = fmt.Sprintf("Failed to invite %s %s (%s) as a guest to '%s': %s", i.firstName(), i.lastName(), i.emailAddress(), i.channel.Name(i.api), err.Error())
		return result, err
	}

	i.logger.Info("successfully-invited-single-channel-user")

	result = fmt.Sprintf("@%s invited %s %s (%s) as a guest to '%s'", i.invitingUser, i.firstName(), i.lastName(), i.emailAddress(), i.channel.Name(i.api))
	return result, nil
}

func (i inviteGuest) AuditMessage() string {
	return fmt.Sprintf(
		"@%s invited %s %s (%s) as a single-channel guest to '%s' (%s)",
		i.invitingUser,
		i.firstName(),
		i.lastName(),
		i.emailAddress(),
		i.channel.Name(i.api),
		i.channel.ID(),
	)
}
