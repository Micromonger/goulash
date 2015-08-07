package action

import "fmt"

const (
	channelNotVisibleErrFmt       = "<@%s> can only invite people to channels or private groups it is a member of. You can invite <@%s> by typing `/invite @%s` from the channel or private group you would like <@%s> to invite people to."
	missingParameterErrFmt        = "Missing required %s parameter. See `%s help` for more information."
	uninvitableDomainErrFmt       = "Users for the '%s' domain are unable to be invited through %s. %s"
	userNotFoundErrFmt            = "Unable to find user matching '%s'."
	fullUserCannotBeErrFmt        = "Full users cannot be %s."
	userIsAlreadyErrFmt           = "User is already a %s."
	cannotFromDirectMessageErrFmt = "Cannot %s from a direct message. Try again from a channel or group."
)

type channelNotVisibleErr struct {
	slackUserID string
}

// NewChannelNotVisibleErr returns an error
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

// NewUninvitableDomainErr returns an error
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

// NewMissingEmailParameterErr returns an error
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

// NewUserNotFoundErr returns an error
func NewUserNotFoundErr(searchParam string) error {
	return userNotFoundErr{
		searchParam: searchParam,
	}
}

func (e userNotFoundErr) Error() string {
	return fmt.Sprintf(userNotFoundErrFmt, e.searchParam)
}

type fullUserCannotBeErr struct {
	verb string
}

// NewFullUserCannotBeErr returns an error
func NewFullUserCannotBeErr(verb string) error {
	return fullUserCannotBeErr{
		verb: verb,
	}
}

func (e fullUserCannotBeErr) Error() string {
	return fmt.Sprintf(fullUserCannotBeErrFmt, e.verb)
}

type userIsAlreadyErr struct {
	noun string
}

// NewUserIsAlreadyErr returns an error
func NewUserIsAlreadyErr(noun string) error {
	return userIsAlreadyErr{
		noun: noun,
	}
}

func (e userIsAlreadyErr) Error() string {
	return fmt.Sprintf(userIsAlreadyErrFmt, e.noun)
}

type cannotFromDirectMessageErr struct {
	verb string
}

// NewCannotFromDirectMessageErr returns an error
func NewCannotFromDirectMessageErr(verb string) error {
	return cannotFromDirectMessageErr{
		verb: verb,
	}
}

func (e cannotFromDirectMessageErr) Error() string {
	return fmt.Sprintf(cannotFromDirectMessageErrFmt, e.verb)
}
