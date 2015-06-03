package config

type localConfig struct {
	slackAuthToken    string
	slackSlashCommand string
	slackTeamName     string
	slackUserID       string

	uninvitableDomain  string
	uninvitableMessage string
	auditLogChannelID  string
}

// NewLocalConfig returns a new Config which will use the provided
// values as its source.
func NewLocalConfig(
	slackAuthToken string,
	slackSlashCommand string,
	slackTeamName string,
	slackUserID string,

	auditLogChannelID string,
	uninvitableDomain string,
	uninvitableMessage string,
) Config {
	return &localConfig{
		slackAuthToken:    slackAuthToken,
		slackSlashCommand: slackSlashCommand,
		slackTeamName:     slackTeamName,
		slackUserID:       slackUserID,

		auditLogChannelID:  auditLogChannelID,
		uninvitableDomain:  uninvitableDomain,
		uninvitableMessage: uninvitableMessage,
	}
}

func (c localConfig) AuditLogChannelID() string {
	return c.auditLogChannelID
}

func (c localConfig) SlackAuthToken() string {
	return c.slackAuthToken
}

func (c localConfig) SlackTeamName() string {
	return c.slackTeamName
}

func (c localConfig) SlackUserID() string {
	return c.slackUserID
}

func (c localConfig) SlackSlashCommand() string {
	return c.slackSlashCommand
}

func (c localConfig) UninvitableDomain() string {
	return c.uninvitableDomain
}

func (c localConfig) UninvitableMessage() string {
	return c.uninvitableMessage
}
