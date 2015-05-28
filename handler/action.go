package handler

import (
	"fmt"

	"github.com/pivotal-golang/lager"
)

// Action represents an action that is able to be performed by the server.
type Action interface {
	Do() error
	SuccessMessage() string
	FailureMessage() string
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

func (i inviteGuestAction) Do() error {
	err := i.api.InviteGuest(
		i.slackTeamName,
		[]string{i.channel.ID},
		i.firstName,
		i.lastName,
		i.emailAddress,
	)
	if err != nil {
		i.logger.Error("failed-inviting-single-channel-user", err)
		return err
	}

	i.logger.Info("successfully-invited-single-channel-user")
	return nil
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

func (i inviteGuestAction) SuccessMessage() string {
	return fmt.Sprintf("@%s invited %s %s (%s) as a guest to '%s'", i.invitingUser, i.firstName, i.lastName, i.emailAddress, i.channel.Name(i.api))
}

func (i inviteGuestAction) FailureMessage() string {
	return fmt.Sprintf("Failed to invite %s %s (%s) as a guest to '%s'", i.firstName, i.lastName, i.emailAddress, i.channel.Name(i.api))
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

func (i inviteRestrictedAction) Do() error {
	err := i.api.InviteRestricted(
		i.slackTeamName,
		i.channel.ID,
		i.firstName,
		i.lastName,
		i.emailAddress,
	)
	if err != nil {
		i.logger.Error("failed-inviting-restricted-account", err)
		return err
	}

	i.logger.Info("successfully-invited-restricted-account")
	return nil
}

func (i inviteRestrictedAction) SuccessMessage() string {
	return fmt.Sprintf("@%s invited %s %s (%s) as a restricted account to '%s'", i.invitingUser, i.firstName, i.lastName, i.emailAddress, i.channel.Name(i.api))
}

func (i inviteRestrictedAction) FailureMessage() string {
	return fmt.Sprintf("Failed to invite %s %s (%s) as a restricted account to '%s'", i.firstName, i.lastName, i.emailAddress, i.channel.Name(i.api))
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
