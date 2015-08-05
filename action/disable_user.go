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

func (du disableUser) searchParam() string {
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
	var message string

	logger = logger.Session("do")

	users, err := api.GetUsers()
	if err != nil {
		logger.Error("failed", err)
		return du.failureMessage(err), err
	}

	var id string
	for _, user := range users {
		if !(user.IsRestricted || user.IsUltraRestricted) {
			err = NewUserCannotBeDisabledErr()
			return du.failureMessage(err), err
		}

		if user.Profile.Email == du.searchParam() || fmt.Sprintf("@%s", user.Name) == du.searchParam() {
			id = user.ID
			break
		}
	}

	if len(id) == 0 {
		logger.Error("failed", err)
		err = NewUserNotFoundErr(du.searchParam())
		return du.failureMessage(err), err
	}

	err = api.DisableUser(config.SlackTeamName(), du.searchParam())
	if err != nil {
		logger.Error("failed", err)
		return du.failureMessage(err), err
	}

	logger.Info("succeeded")
	message = fmt.Sprintf("Successfully disabled user '%s'", du.searchParam())

	return message, nil
}

func (du disableUser) AuditMessage(api slackapi.SlackAPI) string {
	return fmt.Sprintf(
		"@%s disabled user %s",
		du.disablingUser,
		du.searchParam(),
	)
}

func (du disableUser) failureMessage(err error) string {
	return fmt.Sprintf("Failed to disable user '%s': %s", du.searchParam(), err.Error())
}
