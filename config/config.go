package config

// Config is an interface that provides configuration values.
type Config interface {
	AuditLogChannelID() string
	SlackAuthToken() string
	SlackTeamName() string
	SlackUserID() string
	SlackSlashCommand() string
	UninvitableDomain() string
	UninvitableMessage() string
}
