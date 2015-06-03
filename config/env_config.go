package config

import "os"

type envConfig struct {
	slackAuditLogChannelIDVar   string
	slackAuthTokenVar           string
	slackSlashCommandVar        string
	slackTeamNameVar            string
	slackUserIDVar              string
	uninvitableDomainMessageVar string
	uninvitableDomainVar        string
}

// NewEnvConfig returns a new Config which will use environment variables as
// its source.
func NewEnvConfig(
	slackAuditLogChannelIDVar string,
	slackAuthTokenVar string,
	slackSlashCommandVar string,
	slackTeamNameVar string,
	slackUserIDVar string,
	uninvitableDomainMessageVar string,
	uninvitableDomainVar string,
) Config {
	return &envConfig{
		slackAuditLogChannelIDVar:   slackAuditLogChannelIDVar,
		slackAuthTokenVar:           slackAuthTokenVar,
		slackSlashCommandVar:        slackSlashCommandVar,
		slackTeamNameVar:            slackTeamNameVar,
		slackUserIDVar:              slackUserIDVar,
		uninvitableDomainMessageVar: uninvitableDomainMessageVar,
		uninvitableDomainVar:        uninvitableDomainVar,
	}
}

func (c envConfig) AuditLogChannelID() string {
	return os.Getenv(c.slackAuditLogChannelIDVar)
}

func (c envConfig) SlackAuthToken() string {
	return os.Getenv(c.slackAuthTokenVar)
}

func (c envConfig) SlackTeamName() string {
	return os.Getenv(c.slackTeamNameVar)
}

func (c envConfig) SlackUserID() string {
	return os.Getenv(c.slackUserIDVar)
}

func (c envConfig) SlackSlashCommand() string {
	return os.Getenv(c.slackSlashCommandVar)
}

func (c envConfig) UninvitableDomain() string {
	return os.Getenv(c.uninvitableDomainVar)
}

func (c envConfig) UninvitableMessage() string {
	return os.Getenv(c.uninvitableDomainMessageVar)
}
