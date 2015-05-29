package handler

import (
	"errors"
	"fmt"

	"github.com/pivotal-golang/lager"
)

var uninvitableUserNotFoundMessageFmt = "There is no user here with the email address '%s'. %s"
var userInfoMessageFmt = "%s %s (%s) is a Slack %s, with the username <@%s>."
var userNotFoundMessageFmt = "There is no user here with the email address '%s'. You can invite them to Slack as a guest or a restricted account. Type `/butler help` for more information."

var membershipFull = "full member"
var membershipRestrictedAccount = "restricted account"
var membershipSingleChannelGuest = "single-channel guest"

// Action represents an action that is able to be performed by the server.
type Action interface {
	Do() (string, error)
	AuditMessage() string
}

// GuardedAction is an Action with prerequisites described in Check().
type GuardedAction interface {
	Check() error
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

func (i inviteGuestAction) Check() error {
	if !i.channel.Visible(i.api) {
		return channelNotVisibleErr(i.slackUserID)
	}

	return nil
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
	slackUserID   string
	channel       *Channel
	invitingUser  string
	emailAddress  string
	firstName     string
	lastName      string

	logger lager.Logger
}

func (i inviteRestrictedAction) Check() error {
	if !i.channel.Visible(i.api) {
		return channelNotVisibleErr(i.slackUserID)
	}

	return nil
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
	emailAddress             string
	requestingUser           string
	slackTeamName            string
	uninvitableDomainMessage string
	uninvitableDomain        string

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

	if uninvitableEmail(i.emailAddress, i.uninvitableEmail) {
		result = fmt.Sprintf(uninvitableUserNotFoundMessageFmt, i.emailAddress, i.uninvitableMessage)
	} else {
		result = fmt.Sprintf(userNotFoundMessageFmt, i.emailAddress)
	}

	return result, errors.New("user_not_found")
}

func (i userInfoAction) AuditMessage() string {
	return fmt.Sprintf("@%s requested info on '%s'", i.requestingUser, i.emailAddress)
}

func channelNotVisibleErr(slackUserID string) error {
	return fmt.Errorf(
		"<@%s> can only invite people to channels or private groups it is a member of. You can invite <@%s> by typing `/invite @%s` from the channel or private group you would like <@%s> to invite people to.",
		slackUserID,
		slackUserID,
		slackUserID,
		slackUserID,
	)
}
