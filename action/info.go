package action

import (
	"errors"
	"fmt"

	"github.com/pivotal-golang/lager"
	"github.com/pivotalservices/goulash/config"
	"github.com/pivotalservices/goulash/slackapi"
)

var (
	uninvitableUserNotFoundMessageFmt = "There is no user here with the email address '%s'. %s"
	infoMessageFmt                    = "%s %s (%s) is a Slack %s, with the username <@%s>."
	userNotFoundMessageFmt            = "There is no user here with the email address '%s'. You can invite them to Slack as a guest or a restricted account. Type `%s help` for more information."

	membershipFull               = "full member"
	membershipRestrictedAccount  = "restricted account"
	membershipSingleChannelGuest = "single-channel guest"
)

type info struct {
	params         []string
	requestingUser string
}

func (i info) emailAddress() string {
	if len(i.params) >= 0 {
		return i.params[0]
	}

	return ""
}

func (i info) Do(
	config config.Config,
	api slackapi.SlackAPI,
	logger lager.Logger,
) (string, error) {
	var result string

	logger = logger.Session("do")

	users, err := api.GetUsers()
	if err != nil {
		logger.Error("failed-getting-users", err)
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
			logger.Info("successfully-found-user")
			result = fmt.Sprintf(
				infoMessageFmt,
				user.Profile.FirstName,
				user.Profile.LastName,
				user.Profile.Email,
				membership,
				user.Name,
			)
			return result, nil
		}
	}

	if uninvitableEmail(i.emailAddress(), config.UninvitableDomain()) {
		result = fmt.Sprintf(uninvitableUserNotFoundMessageFmt, i.emailAddress(), config.UninvitableMessage())
	} else {
		result = fmt.Sprintf(userNotFoundMessageFmt, i.emailAddress(), config.SlackSlashCommand())
	}

	err = errors.New(result)
	logger.Error("failed-to-find-user", err)

	return result, err
}

func (i info) AuditMessage(api slackapi.SlackAPI) string {
	return fmt.Sprintf("@%s requested info on '%s'", i.requestingUser, i.emailAddress())
}
