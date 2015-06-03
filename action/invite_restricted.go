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
	logger := i.logger.Session("check")

	if uninvitableEmail(i.emailAddress(), i.config.UninvitableDomain()) {
		logger.Info("uninvitable-email", lager.Data{
			"emailAddress":      i.emailAddress(),
			"uninvitableDomain": i.config.UninvitableDomain(),
		})
		return NewUninvitableDomainErr(
			i.config.UninvitableDomain(),
			i.config.UninvitableMessage(),
			i.config.SlackSlashCommand(),
		)
	}

	if !i.channel.Visible(i.api) {
		logger.Info("channel-not-visible", lager.Data{
			"slack_user_id": i.config.SlackUserID(),
			"channelID":     i.channel.ID(),
		})
		return NewChannelNotVisibleErr(i.config.SlackUserID())
	}

	return nil
}

func (i inviteRestricted) Do() (string, error) {
	var result string

	logger := i.logger.Session("do")

	err := i.api.InviteRestricted(
		i.config.SlackTeamName(),
		i.channel.ID(),
		i.firstName(),
		i.lastName(),
		i.emailAddress(),
	)
	if err != nil {
		logger.Error("failed-inviting-restricted-account", err)
		result = fmt.Sprintf("Failed to invite %s %s (%s) as a restricted account to '%s': %s", i.firstName(), i.lastName(), i.emailAddress(), i.channel.Name(i.api), err.Error())
		return result, err
	}

	logger.Info("successfully-invited-restricted-account")

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
