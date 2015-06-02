package action

import "fmt"

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
