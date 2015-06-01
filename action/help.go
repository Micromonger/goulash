package action

type help struct{}

func (h help) Do() (string, error) {
	text := "*USAGE*\n" +
		"`/butler [command] [args]`\n" +
		"\n" +
		"*COMMANDS*\n" +
		"\n" +
		"`invite-guest <email> <firstname> <lastname>`\n" +
		"_Invite a Single-Channel Guest to the current channel/group_\n" +
		"\n" +
		"`invite-restricted <email> <firstname> <lastname>`\n" +
		"_Invite a Restricted Account to the current channel/group_\n" +
		"\n" +
		"`info <email>`\n" +
		"_Get information on a Slack user_\n"

	return text, nil
}
