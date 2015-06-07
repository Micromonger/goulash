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

func (i inviteRestricted) Check(
	config config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) error {
	logger = logger.Session("check")

	if i.emailAddress() == "" {
		logger.Info("missing-email-address")
		return NewMissingEmailParameterErr(config.SlackSlashCommand())
	}

	if uninvitableEmail(i.emailAddress(), config.UninvitableDomain()) {
		logger.Info("uninvitable-email", lager.Data{
			"emailAddress":      i.emailAddress(),
			"uninvitableDomain": config.UninvitableDomain(),
		})
		return NewUninvitableDomainErr(
			config.UninvitableDomain(),
			config.UninvitableMessage(),
			config.SlackSlashCommand(),
		)
	}

	if !i.channel.Visible(api) {
		logger.Info("channel-not-visible", lager.Data{
			"slack_user_id": config.SlackUserID(),
			"channelID":     i.channel.ID(),
		})
		return NewChannelNotVisibleErr(config.SlackUserID())
	}

	return nil
}

func (i inviteRestricted) Do(
	config config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) (string, error) {
	var result string

	logger = logger.Session("do")

	err := api.InviteRestricted(
		config.SlackTeamName(),
		i.channel.ID(),
		i.firstName(),
		i.lastName(),
		i.emailAddress(),
	)
	if err != nil {
		logger.Error("failed-inviting-restricted-account", err)
		result = fmt.Sprintf("Failed to invite %s %s (%s) as a restricted account to '%s': %s", i.firstName(), i.lastName(), i.emailAddress(), i.channel.Name(api), err.Error())
		return result, err
	}

	logger.Info("successfully-invited-restricted-account")

	result = fmt.Sprintf("@%s invited %s %s (%s) as a restricted account to '%s'", i.invitingUser, i.firstName(), i.lastName(), i.emailAddress(), i.channel.Name(api))
	return result, nil
}

func (i inviteRestricted) AuditMessage(api slackapi.SlackAPI) string {
	return fmt.Sprintf(
		"@%s invited %s %s (%s) as a restricted account to '%s' (%s)",
		i.invitingUser,
		i.firstName(),
		i.lastName(),
		i.emailAddress(),
		i.channel.Name(api),
		i.channel.ID(),
	)
}
