package config

import (
	"os"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/pivotal-golang/lager"
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

	logger lager.Logger
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

	logger lager.Logger,
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

		logger: logger,
	}
}

func (c envConfig) AuditLogChannelID() string {
	return os.Getenv(c.slackAuditLogChannelIDVar)
}

func (c envConfig) SlackAuthToken() string {
	logger := c.logger.Session("slack-auth-token")

	if c.configServiceNameVar == "" {
		logger.Info("successfully-found")
		return os.Getenv(c.slackAuthTokenVar)
	}

	configServiceName := os.Getenv(c.configServiceNameVar)
	if configServiceName == "" {
		logger.Error("failed-to-find-config-service-name", nil, lager.Data{
			"configServiceName": configServiceName,
		})
		return ""
	}

	if c.app == nil {
		logger.Error("no-app-given", nil)
		return ""
	}

	service, err := c.app.Services.WithName(configServiceName)
	if err != nil {
		logger.Error("failed-to-find-service", nil, lager.Data{
			"configServiceName": configServiceName,
		})
		return ""
	}

	slackAuthToken, ok := service.Credentials[slackAuthTokenCredentialKey]
	if !ok {
		logger.Error("failed-to-find-service-credential", nil, lager.Data{
			"slackAuthTokenCredentialKey": slackAuthTokenCredentialKey,
		})
		return ""
	}

	slackAuthTokenString, ok := slackAuthToken.(string)
	if !ok {
		logger.Error("failed-to-convert-slack-auth-token-to-string", nil)
		return ""
	}

	logger.Info("successfully-found")

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
