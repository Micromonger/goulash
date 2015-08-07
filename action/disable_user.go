package action

import (
	"fmt"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
	"github.com/pivotalservices/slack"
)

type disableUser struct {
	params        []string
	disablingUser string
}

func (du disableUser) searchVal() string {
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

	user, err := du.check(du.searchVal(), config, api, logger)
	if err != nil {
		logger.Error("check-failed", err)
		return du.failureMessage(err), err
	}

	err = api.DisableUser(config.SlackTeamName(), user.ID)
	if err != nil {
		logger.Error("failed", err)
		return du.failureMessage(err), err
	}

	logger.Info("succeeded")

	return fmt.Sprintf("Successfully disabled user '%s'", du.searchVal()), nil
}

func (du disableUser) AuditMessage(api slackapi.SlackAPI) string {
	return fmt.Sprintf(
		"@%s disabled user %s",
		du.disablingUser,
		du.searchVal(),
	)
}

func (du disableUser) failureMessage(err error) string {
	return fmt.Sprintf(
		"Failed to disable user '%s': %s",
		du.searchVal(),
		err.Error(),
	)
}

func (du disableUser) check(
	searchVal string,
	config config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) (slack.User, error) {
	logger = logger.Session("check")

	user, err := findUser(searchVal, api)
	if err != nil {
		logger.Error("failed", err)
		return slack.User{}, err
	}

	if !(user.IsRestricted || user.IsUltraRestricted) {
		err = NewFullUserCannotBeErr("disabled")
		logger.Error("failed", err)
		return slack.User{}, err
	}

	return user, nil
}
