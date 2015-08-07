package action

import (
	"fmt"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
)

type help struct{}

func (h help) Do(
	config config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) (string, error) {
	text := fmt.Sprintf(
		"*USAGE*\n"+
			"`%s [command] [args]`\n"+
			"\n"+
			"*COMMANDS*\n"+
			"\n"+
			"`disable-user [email|@username]`\n"+
			"_Disable a Slack user_\n"+
			"\n"+
			"`groups`\n"+
			"_List the groups that @%s is in_\n"+
			"\n"+
			"`guestify [email|@username]`\n"+
			"_Convert a Restricted Account to a Single-Channel Guest_\n"+
			"\n"+
			"`info [email]`\n"+
			"_Get information on a Slack user_\n"+
			"\n"+
			"`invite-guest [email] [firstname] [lastname]`\n"+
			"_Invite a Single-Channel Guest to the current channel/group_\n"+
			"\n"+
			"`invite-restricted [email] [firstname] [lastname]`\n"+
			"_Invite a Restricted Account to the current channel/group_\n"+
			"\n"+
			"`request-access [#channel]`\n"+
			"_Request an invitation to a channel_\n",
		config.SlackSlashCommand(),
		config.SlackUserID(),
	)

	return text, nil
}
