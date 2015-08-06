package action

import (
	"fmt"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
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
	guestifyParams := make([]string, 1)
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

	if err := g.check(config, api, logger); err != nil {
		return g.failureMessage(err), err
	}

	searchVal := g.searchVal()

	user, err := findUser(searchVal, api)
	if err != nil {
		logger.Error("failed", err)
		return g.failureMessage(err), err
	}

	if !(user.IsRestricted || user.IsUltraRestricted) {
		err = errUserCannotBeGuestified
		return g.failureMessage(err), err
	}

	if user.IsUltraRestricted {
		err = errUserIsAlreadyUltraRestricted
		return g.failureMessage(err), err
	}

	err = api.SetUltraRestricted(
		config.SlackTeamName(),
		user.ID,
		g.channel.ID(),
	)
	if err != nil {
		logger.Error("failed", err)
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
	config config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) error {
	logger = logger.Session("check")

	if g.channel.Name(api) == slackapi.DirectMessageGroupName {
		err := errCannotGuestifyFromDirectMessage
		logger.Error("failed", err)
		return err
	}

	logger.Info("passed")

	return nil
}
