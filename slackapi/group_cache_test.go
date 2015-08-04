package slackapi_test

import (
	"errors"
	"time"

	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotalservices/goulash/slackapi"
	"github.com/pivotalservices/slack"

	fakeslackapi "github.com/pivotalservices/goulash/slackapi/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GroupCache", func() {
	var (
		groupCache slackapi.GroupCache
		ttl        time.Duration

		fakeSlackAPI *fakeslackapi.FakeSlackAPI
		fakeClock    *fakeclock.FakeClock
	)

	BeforeEach(func() {
		fakeSlackAPI = &fakeslackapi.FakeSlackAPI{}
		initialTime := time.Date(2014, 1, 31, 10, 59, 53, 124235, time.UTC)
		fakeClock = fakeclock.NewFakeClock(initialTime)

		ttl = time.Second
		groupCache = slackapi.NewGroupCache(fakeSlackAPI, ttl, fakeClock)

	})

	Describe("Groups", func() {
		It("makes a GetGroups request when it's empty", func() {
			_, err := groupCache.Groups()
			Ω(err).NotTo(HaveOccurred())
			Ω(fakeSlackAPI.GetGroupsCallCount()).To(Equal(1))
		})

		It("does not make a GetGroups request when it's not empty and the cache hasn't expired", func() {
			fakeSlackAPI.GetGroupsReturns([]slack.Group{
				{
					Name: "group-1",
				},
				{
					Name: "group-2",
				},
			}, nil)

			_, err := groupCache.Groups()
			Ω(err).NotTo(HaveOccurred())
			Ω(fakeSlackAPI.GetGroupsCallCount()).To(Equal(1))

			fakeClock.Increment(ttl / 2)

			_, err = groupCache.Groups()
			Ω(err).NotTo(HaveOccurred())
			Ω(fakeSlackAPI.GetGroupsCallCount()).To(Equal(1))
		})

		It("makes a GetGroups request when the cache is full and has expired", func() {
			fakeSlackAPI.GetGroupsReturns([]slack.Group{
				{
					Name: "group-1",
				},
				{
					Name: "group-2",
				},
			}, nil)

			_, err := groupCache.Groups()
			Ω(err).NotTo(HaveOccurred())
			Ω(fakeSlackAPI.GetGroupsCallCount()).To(Equal(1))

			fakeClock.Increment(ttl * 2)

			_, err = groupCache.Groups()
			Ω(err).NotTo(HaveOccurred())
			Ω(fakeSlackAPI.GetGroupsCallCount()).To(Equal(2))
		})

		It("falls back to the cache when the cache has expired but GetGroups returns an error", func() {
			groups := []slack.Group{
				{
					Name: "group-1",
				},
				{
					Name: "group-2",
				},
			}

			fakeSlackAPI.GetGroupsReturns(groups, nil)

			_, err := groupCache.Groups()
			Ω(err).NotTo(HaveOccurred())

			fakeClock.Increment(ttl * 2)

			fakeSlackAPI.GetGroupsReturns([]slack.Group{}, errors.New("error"))

			actualGroups, err := groupCache.Groups()
			Ω(err).NotTo(HaveOccurred())
			Ω(actualGroups).To(Equal(groups))
		})
	})
})
