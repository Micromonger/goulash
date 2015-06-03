package config

import (
	"os"

	"github.com/cloudfoundry-community/go-cfenv"
)

const (
	slackAuthTokenCredentialKey = "slack-auth-token"
)

type envConfig struct {
	app                         *cfenv.App
	configServiceNameVar        string
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
	app *cfenv.App,
	configServiceNameVar string,
	slackAuditLogChannelIDVar string,
	slackAuthTokenVar string,
	slackSlashCommandVar string,
	slackTeamNameVar string,
	slackUserIDVar string,
	uninvitableDomainMessageVar string,
	uninvitableDomainVar string,
) Config {
	return &envConfig{
		app:                         app,
		configServiceNameVar:        configServiceNameVar,
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
	if c.configServiceNameVar == "" {
		return os.Getenv(c.slackAuthTokenVar)
	}

	configServiceName := os.Getenv(c.configServiceNameVar)
	if configServiceName == "" {
		return ""
	}

	if c.app == nil {
		return ""
	}

	service, err := c.app.Services.WithName(configServiceName)
	if err != nil {
		return ""
	}

	slackAuthToken, ok := service.Credentials[slackAuthTokenCredentialKey]
	if !ok {
		return ""
	}

	slackAuthTokenString, ok := slackAuthToken.(string)
	if !ok {
		return ""
	}

	return slackAuthTokenString
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
