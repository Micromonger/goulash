package action

import (
	"fmt"
	"strings"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
	"github.com/pivotalservices/slack"
)

type accessRequest struct {
	params        []string
	commanderName string
	commanderID   string
}

// NewAccessRequest returns a new AccessRequest action, used to request an
// invitation to a channel.
func NewAccessRequest(
	params []string,
	commanderName string,
	commanderID string,
) Action {
	accessRequestParams := make([]string, 3)
	for i := range params {
		accessRequestParams[i] = params[i]
	}

	return &accessRequest{
		params:        accessRequestParams,
		commanderName: commanderName,
		commanderID:   commanderID,
	}
}

func (a accessRequest) channelName() string {
	name := a.params[0]

	if strings.HasPrefix(name, "#") {
		return name[1:]
	}

	return name
}

func (a accessRequest) Do(
	config config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) (string, error) {
	logger = logger.Session("do")

	err := a.check(api, logger)
	if err != nil {
		logger.Error("failed", err)
		return a.failureMessage(err), err
	}

	channel, err := findChannel(a.channelName(), api)
	if err != nil {
		logger.Error("failed", err)
		return a.failureMessage(err), err
	}

	message := fmt.Sprintf(
		"@%s would like to be invited to this channel. To invite them, use `/invite @%s`",
		a.commanderName,
		a.commanderName,
	)

	postMessageParams := slack.NewPostMessageParameters()
	postMessageParams.AsUser = true
	postMessageParams.Parse = "full"

	_, _, err = api.PostMessage(channel.ID, message, postMessageParams)
	if err != nil {
		logger.Error("failed", err)
		return a.failureMessage(err), err
	}

	logger.Info("succeeded")

	successMessage := fmt.Sprintf(
		"Successfully requested access to <#%s>.",
		a.channelName(),
	)

	return successMessage, nil
}

func (a accessRequest) check(
	api slackapi.SlackAPI,
	logger lager.Logger,
) error {
	logger = logger.Session("check")

	user, err := api.GetUserInfo(a.commanderID)
	if err != nil {
		logger.Error("failed", err)
		return err
	}

	if user.IsRestricted || user.IsUltraRestricted {
		logger.Error("failed", errUnauthorized)
		return errUnauthorized
	}

	logger.Info("passed")

	return nil
}

func (a accessRequest) failureMessage(err error) string {
	return fmt.Sprintf(
		"Failed to request access to #%s: %s",
		a.channelName(),
		err.Error(),
	)
}

func (a accessRequest) AuditMessage(
	api slackapi.SlackAPI,
) string {
	return fmt.Sprintf(
		"@%s requested access to #%s",
		a.commanderName,
		a.channelName(),
	)
}
