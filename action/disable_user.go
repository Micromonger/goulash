package action

import (
	"fmt"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
)

type disableUser struct {
	params        []string
	disablingUser string
}

func (du disableUser) emailAddress() string {
	return du.params[0]
}

// NewDisableUser returns a new disable user action
func NewDisableUser(params []string, disablingUser string) Action {
	disableUserParams := make([]string, 1)
	for i := range params {
		disableUserParams[i] = params[i]
	}

	return &disableUser{
		params:        disableUserParams,
		disablingUser: disablingUser,
	}
}

func (du disableUser) Do(
	config config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) (string, error) {
	logger = logger.Session("do")

	err := api.DisableUser(config.SlackTeamName(), du.emailAddress())
	if err != nil {
		logger.Error("failed", err)
		return du.failureMessage(api, err), err
	}

	logger.Info("succeeded")

	return du.successMessage(api), nil
}

func (du disableUser) successMessage(api slackapi.SlackAPI) string {
	return fmt.Sprintf(
		"Successfully disabled %s",
		du.emailAddress(),
	)
}

func (du disableUser) failureMessage(api slackapi.SlackAPI, err error) string {
	return fmt.Sprintf("Failed to disable %s: %s", du.emailAddress(), err.Error())
}

func (du disableUser) AuditMessage(api slackapi.SlackAPI) string {
	return fmt.Sprintf(
		"@%s disabled user %s",
		du.disablingUser,
		du.emailAddress(),
	)
}

// func (i invite) check(
// 	config config.Config,
// 	api slackapi.SlackAPI,
// 	logger lager.Logger,
// ) error {
// 	logger = logger.Session("check")

// 	if i.emailAddress() == "" {
// 		logger.Info("missing-email-address")
// 		return NewMissingEmailParameterErr(config.SlackSlashCommand())
// 	}

// 	if uninvitableEmail(i.emailAddress(), config.UninvitableDomain()) {
// 		logger.Info("uninvitable-email", lager.Data{
// 			"emailAddress":      i.emailAddress(),
// 			"uninvitableDomain": config.UninvitableDomain(),
// 		})
// 		return NewUninvitableDomainErr(
// 			config.UninvitableDomain(),
// 			config.UninvitableMessage(),
// 			config.SlackSlashCommand(),
// 		)
// 	}

// 	if !i.channel.Visible(api) {
// 		logger.Info("channel-not-visible", lager.Data{
// 			"slack_user_id": config.SlackUserID(),
// 			"channelID":     i.channel.ID(),
// 		})
// 		return NewChannelNotVisibleErr(config.SlackUserID())
// 	}

// 	logger.Info("passed")

// 	return nil
// }
