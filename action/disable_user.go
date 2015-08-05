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
	var message string

	logger = logger.Session("do")

	users, err := api.GetUsers()
	if err != nil {
		logger.Error("failed", err)
		return du.failureMessage(err), err
	}

	searchVal := du.searchVal()

	var id string
	for _, user := range users {
		if user.Profile.Email == searchVal || fmt.Sprintf("@%s", user.Name) == searchVal {
			if !(user.IsRestricted || user.IsUltraRestricted) {
				err = NewUserCannotBeDisabledErr()
				return du.failureMessage(err), err
			}

			id = user.ID
			break
		}
	}

	if len(id) == 0 {
		logger.Error("failed", err)
		err = NewUserNotFoundErr(searchVal)
		return du.failureMessage(err), err
	}

	err = api.DisableUser(config.SlackTeamName(), searchVal)
	if err != nil {
		logger.Error("failed", err)
		return du.failureMessage(err), err
	}

	logger.Info("succeeded")
	message = fmt.Sprintf("Successfully disabled user '%s'", searchVal)

	return message, nil
}

func (du disableUser) AuditMessage(api slackapi.SlackAPI) string {
	return fmt.Sprintf(
		"@%s disabled user %s",
		du.disablingUser,
		du.searchVal(),
	)
}

func (du disableUser) failureMessage(err error) string {
	return fmt.Sprintf("Failed to disable user '%s': %s", du.searchVal(), err.Error())
}
