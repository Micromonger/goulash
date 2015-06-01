package action

import (
	"fmt"

	"github.com/pivotal-golang/lager"
)

type inviteRestricted struct {
	params []string

	api                SlackAPI
	channel            *Channel
	invitingUser       string
	slackTeamName      string
	slackUserID        string
	uninvitableDomain  string
	uninvitableMessage string

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
	if uninvitableEmail(i.emailAddress(), i.uninvitableDomain) {
		return fmt.Errorf(uninvitableDomainErrFmt, i.uninvitableDomain, i.uninvitableMessage)
	}

	if !i.channel.Visible(i.api) {
		return channelNotVisibleErr(i.slackUserID)
	}

	return nil
}

func (i inviteRestricted) Do() (string, error) {
	var result string

	err := i.api.InviteRestricted(
		i.slackTeamName,
		i.channel.ID,
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
		i.channel.ID,
	)
}
