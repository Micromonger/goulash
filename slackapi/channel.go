package slackapi

type channel struct {
	rawName string
	name    string
	id      string
}

//go:generate counterfeiter . Channel

// Channel represents a Channel or Private Group in Slack.
type Channel interface {
	Name(SlackAPI) string
	Visible(SlackAPI) bool
	ID() string
}

// NewChannel returns a new Channel.
func NewChannel(rawName string, id string) Channel {
	return &channel{
		rawName: rawName,
		id:      id,
	}
}

// Name will attempt to retrieve a Private Group's real name from Slack,
// falling back to the name given by Slack otherwise. In the case of Private
// Groups, this is 'privategroup'
func (c *channel) Name(api SlackAPI) string {
	if c.name != "" {
		return c.name
	}

	if c.rawName != PrivateGroupName {
		return c.rawName
	}

	excludeArchived := true
	groups, _ := api.GetGroups(excludeArchived)

	for _, group := range groups {
		if group.BaseChannel.Id == c.id {
			c.name = group.Name
			return c.name
		}
	}

	c.name = c.rawName
	return c.name
}

// Visible returns true if the account associated with the configured
// SLACK_AUTH_TOKEN is a member of the channel, and false if not.
func (c *channel) Visible(api SlackAPI) bool {
	return c.Name(api) != PrivateGroupName
}

// ID returns the channel's ID
func (c *channel) ID() string {
	return c.id
}
