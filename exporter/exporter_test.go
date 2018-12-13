package exporter_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/go-cfclient"

	"github.com/alphagov/paas-prometheus-exporter/exporter"
	"github.com/alphagov/paas-prometheus-exporter/exporter/mocks"
)

var _ = Describe("CheckForNewApps", func() {

	var fakeClient *mocks.FakeCFClient
	var fakeWatcherManager *mocks.FakeWatcherManager
	var ctx context.Context
	var cancel context.CancelFunc

	BeforeEach(func() {
		fakeClient = &mocks.FakeCFClient{}
		fakeWatcherManager = &mocks.FakeWatcherManager{}
		ctx, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		cancel()
	})

	It("creates a new app", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{
			{Guid: "33333333-3333-3333-3333-333333333333", Instances: 1, Name: "foo", State: "STARTED"},
		}, nil)

		e := exporter.New(fakeClient, fakeWatcherManager)

		go e.Start(ctx, 100*time.Millisecond)

		// Assert addWatcher is called once and only once for example not in subsequent runs of `checkForNewApps`
		Eventually(fakeWatcherManager.AddWatcherCallCount).Should(Equal(1))
		Consistently(fakeWatcherManager.AddWatcherCallCount, 200*time.Millisecond).Should(Equal(1))
	})

	It("does not create a new appWatcher if the app state is stopped", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{
			{Guid: "33333333-3333-3333-3333-333333333333", Instances: 1, Name: "foo", State: "STOPPED"},
		}, nil)

		e := exporter.New(fakeClient, fakeWatcherManager)

		go e.Start(ctx, 100*time.Millisecond)

		// Assert addWatcher is not called including in subsequent runs of `checkForNewApps`
		Consistently(fakeWatcherManager.AddWatcherCallCount, 200*time.Millisecond).Should(Equal(0))
	})

	It("creates a new appWatcher if a stopped app is started", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{
			{Guid: "33333333-3333-3333-3333-333333333333", Instances: 1, Name: "foo", State: "STOPPED"},
		}, nil)
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{
			{Guid: "33333333-3333-3333-3333-333333333333", Instances: 1, Name: "foo", State: "STARTED"},
		}, nil)

		e := exporter.New(fakeClient, fakeWatcherManager)

		go e.Start(ctx, 100*time.Millisecond)

		// Assert addWatcher is called once and only once for example not in subsequent runs of `checkForNewApps`
		Eventually(fakeWatcherManager.AddWatcherCallCount).Should(Equal(1))
		Consistently(fakeWatcherManager.AddWatcherCallCount, 200*time.Millisecond).Should(Equal(1))
	})

	It("deletes an AppWatcher when an app is deleted", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{
			{Guid: "33333333-3333-3333-3333-333333333333", Instances: 1, Name: "foo", State: "STARTED"},
		}, nil)
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{}, nil)

		e := exporter.New(fakeClient, fakeWatcherManager)

		go e.Start(ctx, 100*time.Millisecond)

		// Assert addWatcher is called once and only once for example not in subsequent runs of `checkForNewApps`
		Eventually(fakeWatcherManager.AddWatcherCallCount).Should(Equal(1))
		Consistently(fakeWatcherManager.AddWatcherCallCount, 200*time.Millisecond).Should(Equal(1))

		// Assert deleteWatcher is called once and only once for example not in subsequent runs of `checkForNewApps`
		Eventually(fakeWatcherManager.DeleteWatcherCallCount).Should(Equal(1))
		Consistently(fakeWatcherManager.DeleteWatcherCallCount, 200*time.Millisecond).Should(Equal(1))
	})

	It("deletes an AppWatcher when an app is stopped", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{
			{Guid: "11111111-11111-11111-1111-111-11-1-1-1", Instances: 1, Name: "foo", State: "STARTED"},
		}, nil)
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{
			{Guid: "11111111-11111-11111-1111-111-11-1-1-1", Instances: 1, Name: "foo", State: "STOPPED"},
		}, nil)

		e := exporter.New(fakeClient, fakeWatcherManager)

		go e.Start(ctx, 100*time.Millisecond)

		// Assert addWatcher is called once and only once for example not in subsequent runs of `checkForNewApps`
		Eventually(fakeWatcherManager.AddWatcherCallCount).Should(Equal(1))
		Consistently(fakeWatcherManager.AddWatcherCallCount, 200*time.Millisecond).Should(Equal(1))

		// Assert deleteWatcher is called once and only once for example not in subsequent runs of `checkForNewApps`
		Eventually(fakeWatcherManager.DeleteWatcherCallCount).Should(Equal(1))
		Consistently(fakeWatcherManager.DeleteWatcherCallCount, 200*time.Millisecond).Should(Equal(1))
	})

	It("deletes and recreates an AppWatcher when an app is renamed", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{
			{Guid: "33333333-3333-3333-3333-333333333333", Instances: 1, Name: "foo", State: "STARTED"},
		}, nil)
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{
			{Guid: "33333333-3333-3333-3333-333333333333", Instances: 1, Name: "bar", State: "STARTED"},
		}, nil)

		e := exporter.New(fakeClient, fakeWatcherManager)

		go e.Start(ctx, 100*time.Millisecond)

		// Assert addWatcher is called twice and only twice for example not in subsequent runs of `checkForNewApps`
		Eventually(fakeWatcherManager.AddWatcherCallCount).Should(Equal(2))
		Consistently(fakeWatcherManager.AddWatcherCallCount, 200*time.Millisecond).Should(Equal(2))

		// Assert deleteWatcher is called once and only once for example not in subsequent runs of `checkForNewApps`
		Eventually(fakeWatcherManager.DeleteWatcherCallCount).Should(Equal(1))
		Consistently(fakeWatcherManager.DeleteWatcherCallCount, 200*time.Millisecond).Should(Equal(1))
	})

	It("deletes and recreates an AppWatcher when an app's space is renamed", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{
			{
				Guid:      "33333333-3333-3333-3333-333333333333",
				Instances: 1,
				Name:      "foo",
				State:     "STARTED",
				SpaceData: cfclient.SpaceResource{Entity: cfclient.Space{Name: "spacename"}},
			},
		}, nil)

		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{
			{
				Guid:      "33333333-3333-3333-3333-333333333333",
				Instances: 1,
				Name:      "foo",
				State:     "STARTED",
				SpaceData: cfclient.SpaceResource{Entity: cfclient.Space{Name: "spacenamenew"}},
			},
		}, nil)

		e := exporter.New(fakeClient, fakeWatcherManager)

		go e.Start(ctx, 100*time.Millisecond)

		// Assert addWatcher is called twice and only twice for example not in subsequent runs of `checkForNewApps`
		Eventually(fakeWatcherManager.AddWatcherCallCount).Should(Equal(2))
		Consistently(fakeWatcherManager.AddWatcherCallCount, 200*time.Millisecond).Should(Equal(2))

		// Assert deleteWatcher is called once and only once for example not in subsequent runs of `checkForNewApps`
		Eventually(fakeWatcherManager.DeleteWatcherCallCount).Should(Equal(1))
		Consistently(fakeWatcherManager.DeleteWatcherCallCount, 200*time.Millisecond).Should(Equal(1))
	})

	It("deletes and recreates an AppWatcher when an app's org is renamed", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{
			{
				Guid:      "33333333-3333-3333-3333-333333333333",
				Instances: 1,
				Name:      "foo",
				State:     "STARTED",
				SpaceData: cfclient.SpaceResource{
					Entity: cfclient.Space{
						Name: "spacename",
						OrgData: cfclient.OrgResource{
							Entity: cfclient.Org{Name: "orgname"},
						},
					},
				},
			},
		}, nil)

		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{
			{
				Guid:      "33333333-3333-3333-3333-333333333333",
				Instances: 1,
				Name:      "foo",
				State:     "STARTED",
				SpaceData: cfclient.SpaceResource{
					Entity: cfclient.Space{
						Name: "spacename",
						OrgData: cfclient.OrgResource{
							Entity: cfclient.Org{Name: "orgnamenew"},
						},
					},
				},
			},
		}, nil)

		e := exporter.New(fakeClient, fakeWatcherManager)

		go e.Start(ctx, 100*time.Millisecond)

		// Assert addWatcher is called twice and only twice for example not in subsequent runs of `checkForNewApps`
		Eventually(fakeWatcherManager.AddWatcherCallCount).Should(Equal(2))
		Consistently(fakeWatcherManager.AddWatcherCallCount, 200*time.Millisecond).Should(Equal(2))

		// Assert deleteWatcher is called once and only once for example not in subsequent runs of `checkForNewApps`
		Eventually(fakeWatcherManager.DeleteWatcherCallCount).Should(Equal(1))
		Consistently(fakeWatcherManager.DeleteWatcherCallCount, 200*time.Millisecond).Should(Equal(1))
	})

	It("updates an AppWatcher when an app changes size", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{
			{Guid: "33333333-3333-3333-3333-333333333333", Instances: 1, Name: "foo", State: "STARTED"},
		}, nil)
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{
			{Guid: "33333333-3333-3333-3333-333333333333", Instances: 2, Name: "foo", State: "STARTED"},
		}, nil)

		e := exporter.New(fakeClient, fakeWatcherManager)

		go e.Start(ctx, 100*time.Millisecond)

		Eventually(fakeWatcherManager.UpdateAppInstancesCallCount).Should(Equal(1))

		app := fakeWatcherManager.UpdateAppInstancesArgsForCall(0)
		Expect(app.Guid).To(Equal("33333333-3333-3333-3333-333333333333"))
		Expect(app.Instances).To(Equal(2))
	})
})
