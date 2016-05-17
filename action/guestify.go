package action

import (
	"fmt"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
	"github.com/pivotalservices/slack"
)

type guestify struct {
	params          []string
	channel         slackapi.Channel
	guestifyingUser string
}

func (g guestify) searchVal() string {
	return g.params[0]
}

// NewGuestify returns a new guestify action. This is used to convert a user to
// a Single-Channel Guest.
func NewGuestify(
	params []string,
	channel slackapi.Channel,
	guestifyingUser string,
) Action {
	guestifyParams := []string{""}
	for i := range params {
		guestifyParams[i] = params[i]
	}

	return &guestify{
		params:          guestifyParams,
		channel:         channel,
		guestifyingUser: guestifyingUser,
	}
}

func (g guestify) Do(
	config config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) (string, error) {
	logger = logger.Session("do")

	searchVal := g.searchVal()

	user, err := g.check(searchVal, config, api, logger)
	if err != nil {
		logger.Error("check-failed", err)
		return g.failureMessage(err), err
	}

	err = api.SetUltraRestricted(
		config.SlackTeamName(),
		user.ID,
		g.channel.ID(),
	)
	if err != nil {
		logger.Error("failed-guestifying", err)
		return g.failureMessage(err), err
	}

	return fmt.Sprintf("Successfully guestified user %s", searchVal), nil
}

func (g guestify) failureMessage(err error) string {
	return fmt.Sprintf(
		"Failed to guestify user '%s': %s",
		g.searchVal(),
		err.Error(),
	)
}

func (g guestify) AuditMessage(api slackapi.SlackAPI) string {
	return fmt.Sprintf(
		"@%s guestified user '%s'",
		g.guestifyingUser,
		g.searchVal(),
	)
}

func (g guestify) check(
	searchVal string,
	config config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) (slack.User, error) {
	logger = logger.Session("check")

	if g.channel.Name(api) == slackapi.DirectMessageGroupName {
		return slack.User{}, NewCannotFromDirectMessageErr("guestify")
	}

	user, err := findUser(searchVal, api)
	if err != nil {
		return slack.User{}, err
	}

	if !(user.IsRestricted || user.IsUltraRestricted) {
		return slack.User{}, NewFullUserCannotBeErr("guestified")
	}

	if user.IsUltraRestricted {
		return slack.User{}, NewUserIsAlreadyErr("single-channel guest")
	}

	logger.Info("passed")

	return user, nil
}
