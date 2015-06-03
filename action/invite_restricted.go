package action

import (
	"fmt"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
)

type inviteRestricted struct {
	params       []string
	channel      slackapi.Channel
	invitingUser string

	api    slackapi.SlackAPI
	config config.Config
	logger lager.Logger
}

func (i inviteRestricted) emailAddress() string {
	if len(i.params) > 0 {
		return i.params[0]
	}

	return ""
}

func (i inviteRestricted) firstName() string {
	if len(i.params) > 1 {
		return i.params[1]
	}

	return ""
}

func (i inviteRestricted) lastName() string {
	if len(i.params) > 2 {
		return i.params[2]
	}

	return ""
}

func (i inviteRestricted) Check() error {
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

func (i inviteRestricted) Do() (string, error) {
	var result string

	err := i.api.InviteRestricted(
		i.config.SlackTeamName(),
		i.channel.ID(),
		i.firstName(),
		i.lastName(),
		i.emailAddress(),
	)
	if err != nil {
		i.logger.Error("failed-inviting-restricted-account", err)
		result = fmt.Sprintf("Failed to invite %s %s (%s) as a restricted account to '%s': %s", i.firstName(), i.lastName(), i.emailAddress(), i.channel.Name(i.api), err.Error())
		return result, err
	}

	result = fmt.Sprintf("@%s invited %s %s (%s) as a restricted account to '%s'", i.invitingUser, i.firstName(), i.lastName(), i.emailAddress(), i.channel.Name(i.api))
	return result, nil
}

func (i inviteRestricted) AuditMessage() string {
	return fmt.Sprintf(
		"@%s invited %s %s (%s) as a restricted account to '%s' (%s)",
		i.invitingUser,
		i.firstName(),
		i.lastName(),
		i.emailAddress(),
		i.channel.Name(i.api),
		i.channel.ID(),
	)
}
