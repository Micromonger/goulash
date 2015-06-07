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

func (i inviteGuest) Check(
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

func (i inviteGuest) Do(
	config config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) (string, error) {
	var result string

	logger = logger.Session("do")

	err := api.InviteGuest(
		config.SlackTeamName(),
		i.channel.ID(),
		i.firstName(),
		i.lastName(),
		i.emailAddress(),
	)
	if err != nil {
		logger.Error("failed-inviting-single-channel-guest", err)
		result = fmt.Sprintf("Failed to invite %s %s (%s) as a guest to '%s': %s", i.firstName(), i.lastName(), i.emailAddress(), i.channel.Name(api), err.Error())
		return result, err
	}

	logger.Info("successfully-invited-single-channel-guest")

	result = fmt.Sprintf("@%s invited %s %s (%s) as a guest to '%s'", i.invitingUser, i.firstName(), i.lastName(), i.emailAddress(), i.channel.Name(api))
	return result, nil
}

func (i inviteGuest) AuditMessage(api slackapi.SlackAPI) string {
	return fmt.Sprintf(
		"@%s invited %s %s (%s) as a single-channel guest to '%s' (%s)",
		i.invitingUser,
		i.firstName(),
		i.lastName(),
		i.emailAddress(),
		i.channel.Name(api),
		i.channel.ID(),
	)
}
