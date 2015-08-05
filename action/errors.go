package action

import "fmt"

var (
	channelNotVisibleErrFmt = "<@%s> can only invite people to channels or private groups it is a member of. You can invite <@%s> by typing `/invite @%s` from the channel or private group you would like <@%s> to invite people to."
	missingParameterErrFmt  = "Missing required %s parameter. See `%s help` for more information."
	uninvitableDomainErrFmt = "Users for the '%s' domain are unable to be invited through %s. %s"
	userNotFoundErrFmt      = "Unable to find user matching '%s'."
)

// ChannelNotVisibleErr is an error.
type ChannelNotVisibleErr struct {
	slackUserID string
}

// NewChannelNotVisibleErr returns a new ChannelNotVisibleErr.
func NewChannelNotVisibleErr(slackUserID string) ChannelNotVisibleErr {
	return ChannelNotVisibleErr{
		slackUserID: slackUserID,
	}
}

func (e ChannelNotVisibleErr) Error() string {
	return fmt.Sprintf(channelNotVisibleErrFmt, e.slackUserID, e.slackUserID, e.slackUserID, e.slackUserID)
}

// UninvitableDomainErr is an error.
type UninvitableDomainErr struct {
	uninvitableDomain  string
	uninvitableMessage string
	slackSlashCommand  string
}

// NewUninvitableDomainErr returns a new UninvitableDomainErr.
func NewUninvitableDomainErr(
	uninvitableDomain string,
	uninvitableMessage string,
	slackSlashCommand string,
) UninvitableDomainErr {
	return UninvitableDomainErr{
		uninvitableDomain:  uninvitableDomain,
		uninvitableMessage: uninvitableMessage,
		slackSlashCommand:  slackSlashCommand,
	}
}

func (e UninvitableDomainErr) Error() string {
	return fmt.Sprintf(uninvitableDomainErrFmt, e.uninvitableDomain, e.slackSlashCommand, e.uninvitableMessage)
}

// MissingEmailParameterErr is an error.
type MissingEmailParameterErr struct {
	slackSlashCommand string
}

// NewMissingEmailParameterErr returns a new MissingEmailParameterErr.
func NewMissingEmailParameterErr(slackSlashCommand string) MissingEmailParameterErr {
	return MissingEmailParameterErr{
		slackSlashCommand: slackSlashCommand,
	}
}

func (e MissingEmailParameterErr) Error() string {
	return fmt.Sprintf(missingParameterErrFmt, "email address", e.slackSlashCommand)
}

// UserNotFoundErr is an error.
type UserNotFoundErr struct {
	searchParam string
}

// NewUserNotFoundErr returns a new UserNotFoundErr.
func NewUserNotFoundErr(searchParam string) UserNotFoundErr {
	return UserNotFoundErr{
		searchParam: searchParam,
	}
}

func (e UserNotFoundErr) Error() string {
	return fmt.Sprintf(userNotFoundErrFmt, e.searchParam)
}

// UserCannotBeDisabledErr is an error.
type UserCannotBeDisabledErr struct{}

// NewUserCannotBeDisabledErr returns a new UserCannotBeDisabledErr.
func NewUserCannotBeDisabledErr() UserCannotBeDisabledErr {
	return UserCannotBeDisabledErr{}
}

func (e UserCannotBeDisabledErr) Error() string {
	return "Full users cannot be disabled."
}
