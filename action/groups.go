package action

import (
	"fmt"
	"sort"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
	"github.com/pivotalservices/slack"
)

type groups struct {
	commanderName string
}

// NewGroups returns a new Groups action, used to list the groups the user with
// the configured token is in
func NewGroups(commanderName string) Action {
	return &groups{commanderName: commanderName}
}

func (g groups) Do(
	c config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) (string, error) {
	logger = logger.Session("do")

	excludeArchived := true
	groups, err := api.GetGroups(excludeArchived)
	if err != nil {
		logger.Error("failed", err)
		return failureMessage(err), err
	}

	var groupNames []string
	for _, group := range groups {
		groupNames = append(groupNames, group.Name)
	}

	sort.Strings(groupNames)

	messageText := "I'm in the following groups:\n"
	for _, groupName := range groupNames {
		messageText = messageText + fmt.Sprintf("\n%s", groupName)
	}

	postMessageParams := slack.NewPostMessageParameters()
	postMessageParams.AsUser = true

	_, _, err = api.PostMessage(g.commanderName, messageText, postMessageParams)
	if err != nil {
		logger.Error("failed", err)
		return failureMessage(err), err
	}

	logger.Info("succeeded")

	return "Successfully sent a list of the groups I'm in as a direct message.", nil
}

func (g groups) AuditMessage(api slackapi.SlackAPI) string {
	return fmt.Sprintf("@%s requested the groups that I'm in", g.commanderName)
}

func failureMessage(err error) string {
	return fmt.Sprintf("Failed to list the groups I'm in: %s", err.Error())
}
