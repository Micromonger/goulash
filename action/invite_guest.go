package action

import (
	"fmt"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/slackapi"
)

type inviteGuest struct {
	params []string

	api                slackapi.SlackAPI
	channel            slackapi.Channel
	invitingUser       string
	slackTeamName      string
	slackUserID        string
	uninvitableDomain  string
	uninvitableMessage string

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
	if uninvitableEmail(i.emailAddress(), i.uninvitableDomain) {
		return NewUninvitableDomainErr(i.uninvitableDomain, i.uninvitableMessage)
	}

	if !i.channel.Visible(i.api) {
		return NewChannelNotVisibleErr(i.slackUserID)
	}

	return nil
}

func (i inviteGuest) Do() (string, error) {
	var result string

	err := i.api.InviteGuest(
		i.slackTeamName,
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
