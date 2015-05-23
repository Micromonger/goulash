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

type inviteGuestAction struct {
	api           SlackAPI
	slackTeamName string
	channelID     string
	invitingUser  string
	emailAddress  string
	firstName     string
	lastName      string

	logger lager.Logger
}

func (i inviteGuestAction) Do() error {
	err := i.api.InviteGuest(
		i.slackTeamName,
		i.channelID,
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

func (i inviteGuestAction) SuccessMessage() string {
	return fmt.Sprintf("@%s invited %s %s (%s) as a guest to this channel", i.invitingUser, i.firstName, i.lastName, i.emailAddress)
}

func (i inviteGuestAction) FailureMessage() string {
	return fmt.Sprintf("Failed to invite %s %s (%s) as a guest to this channel", i.firstName, i.lastName, i.emailAddress)
}

func (i inviteGuestAction) AuditMessage() string {
	return fmt.Sprintf(
		"@%s invited %s %s (%s) as a single-channel guest to channel with ID %s",
		i.invitingUser,
		i.firstName,
		i.lastName,
		i.emailAddress,
		i.channelID,
	)
}
