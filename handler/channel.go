package handler

// Channel represents a Channel or Private Group in Slack.
type Channel struct {
	RawName string
	ID      string
}

// Name will attempt to retrieve a Private Group's real name from Slack,
// falling back to the name given by Slack otherwise. In the case of Private
// Groups, this is 'privategroup'
func (c Channel) Name(api SlackAPI) string {
	if c.RawName != "privategroup" {
		return c.RawName
	}

	excludeArchived := true
	groups, _ := api.GetGroups(excludeArchived)

	for _, group := range groups {
		if group.BaseChannel.Id == c.ID {
			return group.Name
		}
	}

	return c.RawName
}
