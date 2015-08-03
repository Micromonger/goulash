package slackapi

import (
	"time"

	"github.com/pivotal-golang/clock"
	"github.com/pivotalservices/slack"
)

const excludeArchived = true

type groupCache struct {
	api   SlackAPI
	ttl   time.Duration
	clock clock.Clock

	groups []slack.Group
	expiry time.Time
}

type GroupCache interface {
	Groups() ([]slack.Group, error)
}

func NewGroupCache(api SlackAPI, ttl time.Duration, clock clock.Clock) GroupCache {
	return &groupCache{
		api:    api,
		ttl:    ttl,
		clock:  clock,
		groups: []slack.Group{},
	}
}

func (gc *groupCache) Groups() ([]slack.Group, error) {
	// add mutex!
	//
	//
	//

	if len(gc.groups) > 0 {
		if !gc.expired() {
			return gc.groups, nil
		}
	}

	groups, _ := gc.api.GetGroups(excludeArchived)

	gc.groups = groups
	gc.expiry = gc.clock.Now().Add(gc.ttl)

	return groups, nil
}

func (gc *groupCache) expired() bool {
	now := gc.clock.Now()

	return now.Add(gc.ttl).After(gc.expiry)
}
