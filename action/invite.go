package action

import (
	"fmt"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
)

type invite struct {
	params       []string
	command      string
	channel      slackapi.Channel
	invitingUser string
}

func (i invite) emailAddress() string {
	if len(i.params) > 0 {
		return i.params[0]
	}

	return ""
}

func (i invite) firstName() string {
	if len(i.params) > 1 {
		return i.params[1]
	}

	return ""
}

func (i invite) lastName() string {
	if len(i.params) > 2 {
		return i.params[2]
	}

	return ""
}

func (i invite) inviteeType() string {
	switch i.command {
	case "invite-guest":
		return "single-channel guest"
	case "invite-restricted":
		return "restricted account"
	}

	return "unknown"
}

func (i invite) Do(
	config config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) (string, error) {
	logger = logger.Session("do")

	if checkErr := i.check(config, api, logger); checkErr != nil {
		return checkErr.Error(), checkErr
	}

	var inviteErr error
	switch i.command {
	case "invite-guest":
		inviteErr = api.InviteGuest(
			config.SlackTeamName(),
			i.channel.ID(),
			i.firstName(),
			i.lastName(),
			i.emailAddress(),
		)

	case "invite-restricted":
		inviteErr = api.InviteRestricted(
			config.SlackTeamName(),
			i.channel.ID(),
			i.firstName(),
			i.lastName(),
			i.emailAddress(),
		)
	}

	if inviteErr != nil {
		logger.Error("failed", inviteErr)
		return i.failureMessage(api, inviteErr), inviteErr
	}

	logger.Info("succeeded")

	return i.successMessage(api), nil
}

func (i invite) AuditMessage(api slackapi.SlackAPI) string {
	return fmt.Sprintf(
		"@%s invited %s %s (%s) as a %s to '%s' (%s)",
		i.invitingUser,
		i.firstName(),
		i.lastName(),
		i.emailAddress(),
		i.inviteeType(),
		i.channel.Name(api),
		i.channel.ID(),
	)
}

func (i invite) successMessage(api slackapi.SlackAPI) string {
	return fmt.Sprintf(
		"@%s invited %s %s (%s) as a %s to '%s'",
		i.invitingUser,
		i.firstName(),
		i.lastName(),
		i.emailAddress(),
		i.inviteeType(),
		i.channel.Name(api),
	)
}

func (i invite) failureMessage(api slackapi.SlackAPI, err error) string {
	return fmt.Sprintf(
		"Failed to invite %s %s (%s) as a %s to '%s': %s",
		i.firstName(),
		i.lastName(),
		i.emailAddress(),
		i.inviteeType(),
		i.channel.Name(api),
		err.Error(),
	)
}

func (i invite) check(
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

	logger.Info("passed")

	return nil
}
