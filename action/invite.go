package action

import (
	"fmt"
	"regexp"

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

// NewInvite returns a new invite action
func NewInvite(
	params []string,
	command string,
	channel slackapi.Channel,
	invitingUser string,
) Action {
	inviteParams := []string{"", "", ""}
	for i := range params {
		inviteParams[i] = params[i]
	}

	return &invite{
		params:       inviteParams,
		command:      command,
		channel:      channel,
		invitingUser: invitingUser,
	}
}

func (i invite) Do(
	config config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) (string, error) {
	var err error

	logger = logger.Session("do")

	if err = i.check(config, api, logger); err != nil {
		return err.Error(), err
	}

	switch i.command {
	case "invite-guest":
		err = api.InviteGuest(
			config.SlackTeamName(),
			i.channel.ID(),
			i.firstName(),
			i.lastName(),
			i.emailAddress(),
		)

	case "invite-restricted":
		err = api.InviteRestricted(
			config.SlackTeamName(),
			i.channel.ID(),
			i.firstName(),
			i.lastName(),
			i.emailAddress(),
		)
	}

	if err != nil {
		alreadyInvited, matchErr := regexp.MatchString("already_invited", err.Error())
		if matchErr != nil {
			return i.failureMessage(api, matchErr), matchErr
		}
		if alreadyInvited {
			return i.successMessage(api), nil
		}

		logger.Error("failed", err)
		return i.failureMessage(api, err), err
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
		"Successfully invited %s %s (%s) as a %s to '%s'",
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

func (i invite) emailAddress() string {
	return i.params[0]
}

func (i invite) firstName() string {
	return i.params[1]
}

func (i invite) lastName() string {
	return i.params[2]
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
