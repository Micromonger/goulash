package slackapi

import (
	"sync"
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
	mutex  *sync.Mutex
}

type GroupCache interface {
	Groups() ([]slack.Group, error)
}

func NewGroupCache(
	api SlackAPI,
	ttl time.Duration,
	clock clock.Clock,
) GroupCache {
	return &groupCache{
		api:   api,
		ttl:   ttl,
		clock: clock,

		groups: []slack.Group{},
		mutex:  &sync.Mutex{},
	}
}

func (gc *groupCache) Groups() ([]slack.Group, error) {
	gc.mutex.Lock()
	defer gc.mutex.Unlock()

	if len(gc.groups) > 0 && !gc.expired() {
		return gc.groups, nil
	}

	groups, err := gc.api.GetGroups(excludeArchived)
	if err != nil && len(gc.groups) > 0 {
		return gc.groups, nil
	}

	gc.groups = groups
	gc.expiry = gc.clock.Now().Add(gc.ttl)

	return groups, nil
}

func (gc *groupCache) expired() bool {
	return gc.clock.Now().After(gc.expiry)
}
