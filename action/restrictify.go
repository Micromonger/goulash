package action

import (
	"fmt"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
	"github.com/pivotalservices/slack"
)

type restrictify struct {
	params          []string
	channel         slackapi.Channel
	restrictingUser string
}

func (r restrictify) searchVal() string {
	return r.params[0]
}

// NewRestrictify returns a new restrictify action. This is used to convert a
// user to a Restricted Account.
func NewRestrictify(
	params []string,
	channel slackapi.Channel,
	restrictingUser string,
) Action {
	restrictifyParams := make([]string, 1)
	for i := range params {
		restrictifyParams[i] = params[i]
	}

	return &restrictify{
		params:          restrictifyParams,
		channel:         channel,
		restrictingUser: restrictingUser,
	}
}

func (r restrictify) Do(
	config config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) (string, error) {
	logger = logger.Session("do")

	searchVal := r.searchVal()

	user, err := r.check(searchVal, config, api, logger)
	if err != nil {
		logger.Error("check-failed", err)
		return r.failureMessage(err), err
	}

	err = api.SetRestricted(config.SlackTeamName(), user.ID)
	if err != nil {
		logger.Error("failed-restrictifying", err)
		return r.failureMessage(err), err
	}

	return fmt.Sprintf("Successfully restrictified user %s", searchVal), nil
}

func (r restrictify) failureMessage(err error) string {
	return fmt.Sprintf(
		"Failed to restrictify user '%s': %s",
		r.searchVal(),
		err.Error(),
	)
}

func (r restrictify) AuditMessage(api slackapi.SlackAPI) string {
	return fmt.Sprintf(
		"@%s restrictified user '%s'",
		r.restrictingUser,
		r.searchVal(),
	)
}

func (r restrictify) check(
	searchVal string,
	config config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) (slack.User, error) {
	logger = logger.Session("check")

	if r.channel.Name(api) == slackapi.DirectMessageGroupName {
		return slack.User{}, NewCannotFromDirectMessageErr("restrictify")
	}

	user, err := findUser(searchVal, api)
	if err != nil {
		return slack.User{}, err
	}

	if !(user.IsRestricted || user.IsUltraRestricted) {
		return slack.User{}, NewFullUserCannotBeErr("restrictified")
	}

	if user.IsRestricted {
		return slack.User{}, NewUserIsAlreadyErr("restricted account")
	}

	logger.Info("passed")

	return user, nil
}
