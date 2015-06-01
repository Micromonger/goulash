package action

import (
	"errors"
	"fmt"

	"github.com/pivotal-golang/lager"
)

type userInfo struct {
	params []string

	api                SlackAPI
	requestingUser     string
	slackTeamName      string
	uninvitableDomain  string
	uninvitableMessage string
	logger             lager.Logger
}

func (i userInfo) emailAddress() string {
	if len(i.params) >= 0 {
		return i.params[0]
	}

	return ""
}

func (i userInfo) Do() (string, error) {
	var result string

	users, err := i.api.GetUsers()
	if err != nil {
		i.logger.Error("failed-getting-users", err)
		result = fmt.Sprintf("Failed to look up user@example.com: %s", err.Error())
		return result, err
	}

	for _, user := range users {
		if user.Profile.Email == i.emailAddress() {
			membership := membershipFull
			if user.IsRestricted {
				membership = membershipRestrictedAccount
			}
			if user.IsUltraRestricted {
				membership = membershipSingleChannelGuest
			}
			result = fmt.Sprintf(
				userInfoMessageFmt,
				user.Profile.FirstName,
				user.Profile.LastName,
				user.Profile.Email,
				membership,
				user.Name,
			)
			return result, nil
		}
	}

	if uninvitableEmail(i.emailAddress(), i.uninvitableDomain) {
		result = fmt.Sprintf(uninvitableUserNotFoundMessageFmt, i.emailAddress(), i.uninvitableMessage)
	} else {
		result = fmt.Sprintf(userNotFoundMessageFmt, i.emailAddress())
	}

	return result, errors.New("user_not_found")
}

func (i userInfo) AuditMessage() string {
	return fmt.Sprintf("@%s requested info on '%s'", i.requestingUser, i.emailAddress())
}
