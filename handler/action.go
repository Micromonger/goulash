package handler

import (
	"errors"
	"fmt"

	"github.com/pivotal-golang/lager"
)

// Action represents an action that is able to be performed by the server.
type Action interface {
	Do() (string, error)
	AuditMessage() string
}

// GuardedAction is an Action with prerequisites. Use Guard() to return whether
// the prerequisite(s) for the action are met, and GuardMessage() for the
// message when the prerequisite(s) are not met.
type GuardedAction interface {
	Guard() bool
	GuardMessage() string
}

type inviteGuestAction struct {
	api           SlackAPI
	slackTeamName string
	slackUserID   string
	channel       *Channel
	invitingUser  string
	emailAddress  string
	firstName     string
	lastName      string

	logger lager.Logger
}

func (i inviteGuestAction) Guard() bool {
	return !i.channel.Visible(i.api)
}

func (i inviteGuestAction) Do() (string, error) {
	var result string

	err := i.api.InviteGuest(
		i.slackTeamName,
		[]string{i.channel.ID},
		i.firstName,
		i.lastName,
		i.emailAddress,
	)
	if err != nil {
		i.logger.Error("failed-inviting-single-channel-user", err)
		result = fmt.Sprintf("Failed to invite %s %s (%s) as a guest to '%s': %s", i.firstName, i.lastName, i.emailAddress, i.channel.Name(i.api), err.Error())
		return result, err
	}

	i.logger.Info("successfully-invited-single-channel-user")

	result = fmt.Sprintf("@%s invited %s %s (%s) as a guest to '%s'", i.invitingUser, i.firstName, i.lastName, i.emailAddress, i.channel.Name(i.api))
	return result, nil
}

func (i inviteGuestAction) GuardMessage() string {
	return fmt.Sprintf(
		"<@%s> can only invite people to channels it is a member of. You can invite <@%s> by typing `/invite @%s` from the channel you would like <@%s> to invite people to.",
		i.slackUserID,
		i.slackUserID,
		i.slackUserID,
		i.slackUserID,
	)
}

func (i inviteGuestAction) AuditMessage() string {
	return fmt.Sprintf(
		"@%s invited %s %s (%s) as a single-channel guest to '%s' (%s)",
		i.invitingUser,
		i.firstName,
		i.lastName,
		i.emailAddress,
		i.channel.Name(i.api),
		i.channel.ID,
	)
}

type inviteRestrictedAction struct {
	api           SlackAPI
	slackTeamName string
	channel       *Channel
	invitingUser  string
	emailAddress  string
	firstName     string
	lastName      string

	logger lager.Logger
}

func (i inviteRestrictedAction) Do() (string, error) {
	var result string

	err := i.api.InviteRestricted(
		i.slackTeamName,
		i.channel.ID,
		i.firstName,
		i.lastName,
		i.emailAddress,
	)
	if err != nil {
		i.logger.Error("failed-inviting-restricted-account", err)
		result = fmt.Sprintf("Failed to invite %s %s (%s) as a restricted account to '%s': %s", i.firstName, i.lastName, i.emailAddress, i.channel.Name(i.api), err.Error())
		return result, err
	}

	result = fmt.Sprintf("@%s invited %s %s (%s) as a restricted account to '%s'", i.invitingUser, i.firstName, i.lastName, i.emailAddress, i.channel.Name(i.api))
	return result, nil
}

func (i inviteRestrictedAction) AuditMessage() string {
	return fmt.Sprintf(
		"@%s invited %s %s (%s) as a restricted account to '%s' (%s)",
		i.invitingUser,
		i.firstName,
		i.lastName,
		i.emailAddress,
		i.channel.Name(i.api),
		i.channel.ID,
	)
}

type userInfoAction struct {
	emailAddress   string
	requestingUser string
	slackTeamName  string

	api    SlackAPI
	logger lager.Logger
}

func (i userInfoAction) Do() (string, error) {
	var result string

	users, err := i.api.GetUsers()
	if err != nil {
		i.logger.Error("failed-getting-users", err)
		result = fmt.Sprintf("Failed to look up user@example.com: %s", err.Error())
		return result, err
	}

	for _, user := range users {
		if user.Profile.Email == i.emailAddress {
			membership := "full member"
			if user.IsRestricted {
				membership = "restricted account"
			}
			if user.IsUltraRestricted {
				membership = "single-channel guest"
			}
			result = fmt.Sprintf(
				"%s %s (%s) is a Slack %s, with the username <@%s>.",
				user.Profile.FirstName,
				user.Profile.LastName,
				user.Profile.Email,
				membership,
				user.Name,
			)
			return result, nil
		}
	}

	result = fmt.Sprintf(
		"There is no user here with the email address '%s'. You can invite them to Slack as a guest or a restricted account. Type `/butler help` for more information.",
		i.emailAddress,
	)

	return result, errors.New("user_not_found")
}

func (i userInfoAction) AuditMessage() string {
	return fmt.Sprintf("@%s requested info on '%s'", i.requestingUser, i.emailAddress)
}
