package config

import "os"

const (
	slackAuditLogChannelIDVar   = "SLACK_AUDIT_LOG_CHANNEL_ID"
	slackAuthTokenVar           = "SLACK_AUTH_TOKEN"
	slackSlashCommandVar        = "SLACK_SLASH_COMMAND"
	slackTeamNameVar            = "SLACK_TEAM_NAME"
	slackUserIDVar              = "SLACK_USER_ID"
	uninvitableDomainMessageVar = "UNINVITABLE_DOMAIN_MESSAGE"
	uninvitableDomainVar        = "UNINVITABLE_DOMAIN"
)

type envConfig struct{}

// NewEnvConfig returns a new Config which will use environment variables as
// its source.
func NewEnvConfig() Config {
	return &envConfig{}
}

func (c envConfig) AuditLogChannelID() string {
	return os.Getenv(slackAuditLogChannelIDVar)
}

func (c envConfig) SlackAuthToken() string {
	return os.Getenv(slackAuthTokenVar)
}

func (c envConfig) SlackTeamName() string {
	return os.Getenv(slackTeamNameVar)
}

func (c envConfig) SlackUserID() string {
	return os.Getenv(slackUserIDVar)
}

func (c envConfig) SlackSlashCommand() string {
	return os.Getenv(slackSlashCommandVar)
}

func (c envConfig) UninvitableDomain() string {
	return os.Getenv(uninvitableDomainVar)
}

func (c envConfig) UninvitableMessage() string {
	return os.Getenv(uninvitableDomainMessageVar)
}
