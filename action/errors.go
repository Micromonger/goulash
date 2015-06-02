package action

import "fmt"

var (
	channelNotVisibleErrFmt = "<@%s> can only invite people to channels or private groups it is a member of. You can invite <@%s> by typing `/invite @%s` from the channel or private group you would like <@%s> to invite people to."
	missingParameterErrFmt  = "Missing required %s parameter. See `/butler help` for more information."
	uninvitableDomainErrFmt = "Users for the '%s' domain are unable to be invited through /butler. %s"
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
}

// NewUninvitableDomainErr returns a new UninvitableDomainErr.
func NewUninvitableDomainErr(uninvitableDomain string, uninvitableMessage string) UninvitableDomainErr {
	return UninvitableDomainErr{
		uninvitableDomain:  uninvitableDomain,
		uninvitableMessage: uninvitableMessage,
	}
}

func (e UninvitableDomainErr) Error() string {
	return fmt.Sprintf(uninvitableDomainErrFmt, e.uninvitableDomain, e.uninvitableMessage)
}

// MissingEmailParameterErr is an error.
type MissingEmailParameterErr struct{}

// NewMissingEmailParameterErr returns a new MissingEmailParameterErr.
func NewMissingEmailParameterErr() MissingEmailParameterErr {
	return MissingEmailParameterErr{}
}

func (e MissingEmailParameterErr) Error() string {
	return fmt.Sprintf(missingParameterErrFmt, "email address")
}
