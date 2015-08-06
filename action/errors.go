package action

import (
	"errors"
	"fmt"
)

var (
	channelNotVisibleErrFmt = "<@%s> can only invite people to channels or private groups it is a member of. You can invite <@%s> by typing `/invite @%s` from the channel or private group you would like <@%s> to invite people to."
	missingParameterErrFmt  = "Missing required %s parameter. See `%s help` for more information."
	uninvitableDomainErrFmt = "Users for the '%s' domain are unable to be invited through %s. %s"
	userNotFoundErrFmt      = "Unable to find user matching '%s'."

	errUserCannotBeDisabled = errors.New("Full users cannot be disabled.")
)

type channelNotVisibleErr struct {
	slackUserID string
}

// NewChannelNotVisibleErr returns a new ChannelNotVisibleErr.
func NewChannelNotVisibleErr(slackUserID string) error {
	return channelNotVisibleErr{
		slackUserID: slackUserID,
	}
}

func (e channelNotVisibleErr) Error() string {
	return fmt.Sprintf(channelNotVisibleErrFmt, e.slackUserID, e.slackUserID, e.slackUserID, e.slackUserID)
}

type uninvitableDomainErr struct {
	uninvitableDomain  string
	uninvitableMessage string
	slackSlashCommand  string
}

// NewUninvitableDomainErr returns a new UninvitableDomainErr.
func NewUninvitableDomainErr(
	uninvitableDomain string,
	uninvitableMessage string,
	slackSlashCommand string,
) error {
	return uninvitableDomainErr{
		uninvitableDomain:  uninvitableDomain,
		uninvitableMessage: uninvitableMessage,
		slackSlashCommand:  slackSlashCommand,
	}
}

func (e uninvitableDomainErr) Error() string {
	return fmt.Sprintf(uninvitableDomainErrFmt, e.uninvitableDomain, e.slackSlashCommand, e.uninvitableMessage)
}

type missingEmailParameterErr struct {
	slackSlashCommand string
}

// NewMissingEmailParameterErr returns a new MissingEmailParameterErr.
func NewMissingEmailParameterErr(slackSlashCommand string) error {
	return missingEmailParameterErr{
		slackSlashCommand: slackSlashCommand,
	}
}

func (e missingEmailParameterErr) Error() string {
	return fmt.Sprintf(missingParameterErrFmt, "email address", e.slackSlashCommand)
}

type userNotFoundErr struct {
	searchParam string
}

// NewUserNotFoundErr returns a new UserNotFoundErr.
func NewUserNotFoundErr(searchParam string) error {
	return userNotFoundErr{
		searchParam: searchParam,
	}
}

func (e userNotFoundErr) Error() string {
	return fmt.Sprintf(userNotFoundErrFmt, e.searchParam)
}
