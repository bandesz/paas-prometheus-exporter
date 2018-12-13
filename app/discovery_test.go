package app_test

import (
	"context"
	"time"

	"github.com/alphagov/paas-prometheus-exporter/app"
	"github.com/alphagov/paas-prometheus-exporter/cf/mocks"
	testmocks "github.com/alphagov/paas-prometheus-exporter/test/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/cloudfoundry-community/go-cfclient"
)

var _ = Describe("CheckForNewApps", func() {

	var discovery *app.Discovery
	var fakeClient *mocks.FakeClient
	var ctx context.Context
	var cancel context.CancelFunc
	var fakeRegisterer *testmocks.FakeRegisterer

	BeforeEach(func() {
		fakeClient = &mocks.FakeClient{}
		fakeRegisterer = &testmocks.FakeRegisterer{}
		discovery = app.NewDiscovery(fakeClient, fakeRegisterer, 100*time.Millisecond)
		ctx, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		cancel()
	})

	It("checks for new apps regularly", func() {
		go discovery.Start(ctx)

		Eventually(fakeClient.ListAppsWithSpaceAndOrgCallCount).Should(Equal(2))
	})

	FIt("creates a new app", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{
			{Guid: "33333333-3333-3333-3333-333333333333", Instances: 1, Name: "foo", State: "STARTED"},
		}, nil)

		fakeAppStreamProvider := &mocks.FakeAppStreamProvider{}
		fakeClient.NewAppStreamProviderReturns(fakeAppStreamProvider)

		go discovery.Start(ctx)

		Eventually(fakeClient.NewAppStreamProviderCallCount).Should(Equal(1))
		Eventually(fakeAppStreamProvider.StartCallCount).Should(Equal(1))

		Consistently(fakeAppStreamProvider.StartCallCount, 200*time.Millisecond).Should(Equal(1))

		Eventually(fakeRegisterer.RegisterCallCount).Should(BeNumerically(">", 0))
	})

	It("does not create a new appWatcher if the app state is stopped", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{
			{Guid: "33333333-3333-3333-3333-333333333333", Instances: 1, Name: "foo", State: "STOPPED"},
		}, nil)

		e := app.NewDiscovery(fakeClient, prometheus.DefaultRegisterer, 100*time.Millisecond)

		go e.Start(ctx)

		Consistently(fakeClient.NewAppStreamProviderCallCount, 200*time.Millisecond).Should(Equal(0))
	})

	It("deletes an AppWatcher when an app is deleted", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{
			{Guid: "33333333-3333-3333-3333-333333333333", Instances: 1, Name: "foo", State: "STARTED"},
		}, nil)
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{}, nil)

		fakeAppStreamProvider := &mocks.FakeAppStreamProvider{}
		fakeClient.NewAppStreamProviderReturns(fakeAppStreamProvider)

		go discovery.Start(ctx)

		Eventually(fakeClient.NewAppStreamProviderCallCount).Should(Equal(1))
		Eventually(fakeAppStreamProvider.StartCallCount).Should(Equal(1))
		Eventually(fakeAppStreamProvider.CloseCallCount).Should(Equal(1))

		Consistently(fakeAppStreamProvider.StartCallCount, 200*time.Millisecond).Should(Equal(1))
		Consistently(fakeAppStreamProvider.CloseCallCount, 200*time.Millisecond).Should(Equal(1))
	})

	It("deletes an AppWatcher when an app is stopped", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{
			{Guid: "11111111-11111-11111-1111-111-11-1-1-1", Instances: 1, Name: "foo", State: "STARTED"},
		}, nil)
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{
			{Guid: "11111111-11111-11111-1111-111-11-1-1-1", Instances: 1, Name: "foo", State: "STOPPED"},
		}, nil)

		fakeAppStreamProvider := &mocks.FakeAppStreamProvider{}
		fakeClient.NewAppStreamProviderReturns(fakeAppStreamProvider)

		go discovery.Start(ctx)

		Eventually(fakeClient.NewAppStreamProviderCallCount).Should(Equal(1))
		Eventually(fakeAppStreamProvider.StartCallCount).Should(Equal(1))
		Eventually(fakeAppStreamProvider.CloseCallCount).Should(Equal(1))

		Consistently(fakeAppStreamProvider.StartCallCount, 200*time.Millisecond).Should(Equal(1))
		Consistently(fakeAppStreamProvider.CloseCallCount, 200*time.Millisecond).Should(Equal(1))
	})

	It("deletes and recreates an AppWatcher when an app is renamed", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{
			{Guid: "33333333-3333-3333-3333-333333333333", Instances: 1, Name: "foo", State: "STARTED"},
		}, nil)
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{
			{Guid: "33333333-3333-3333-3333-333333333333", Instances: 1, Name: "bar", State: "STARTED"},
		}, nil)

		fakeAppStreamProvider1 := &mocks.FakeAppStreamProvider{}
		fakeClient.NewAppStreamProviderReturnsOnCall(0, fakeAppStreamProvider1)
		fakeAppStreamProvider2 := &mocks.FakeAppStreamProvider{}
		fakeClient.NewAppStreamProviderReturnsOnCall(1, fakeAppStreamProvider2)

		go discovery.Start(ctx)

		Eventually(fakeClient.NewAppStreamProviderCallCount).Should(Equal(2))
		Eventually(fakeAppStreamProvider1.StartCallCount).Should(Equal(1))
		Eventually(fakeAppStreamProvider2.StartCallCount).Should(Equal(1))
		Eventually(fakeAppStreamProvider1.CloseCallCount).Should(Equal(1))

		Consistently(fakeAppStreamProvider1.StartCallCount, 200*time.Millisecond).Should(Equal(1))
		Consistently(fakeAppStreamProvider1.CloseCallCount, 200*time.Millisecond).Should(Equal(1))
		Consistently(fakeAppStreamProvider2.StartCallCount, 200*time.Millisecond).Should(Equal(1))
		Consistently(fakeAppStreamProvider2.CloseCallCount, 200*time.Millisecond).Should(Equal(0))
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

		fakeAppStreamProvider1 := &mocks.FakeAppStreamProvider{}
		fakeClient.NewAppStreamProviderReturnsOnCall(0, fakeAppStreamProvider1)
		fakeAppStreamProvider2 := &mocks.FakeAppStreamProvider{}
		fakeClient.NewAppStreamProviderReturnsOnCall(1, fakeAppStreamProvider2)

		go discovery.Start(ctx)

		Eventually(fakeClient.NewAppStreamProviderCallCount).Should(Equal(2))
		Eventually(fakeAppStreamProvider1.StartCallCount).Should(Equal(1))
		Eventually(fakeAppStreamProvider2.StartCallCount).Should(Equal(1))
		Eventually(fakeAppStreamProvider1.CloseCallCount).Should(Equal(1))

		Consistently(fakeAppStreamProvider1.StartCallCount, 200*time.Millisecond).Should(Equal(1))
		Consistently(fakeAppStreamProvider1.CloseCallCount, 200*time.Millisecond).Should(Equal(1))
		Consistently(fakeAppStreamProvider2.StartCallCount, 200*time.Millisecond).Should(Equal(1))
		Consistently(fakeAppStreamProvider2.CloseCallCount, 200*time.Millisecond).Should(Equal(0))
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

		fakeAppStreamProvider1 := &mocks.FakeAppStreamProvider{}
		fakeClient.NewAppStreamProviderReturnsOnCall(0, fakeAppStreamProvider1)
		fakeAppStreamProvider2 := &mocks.FakeAppStreamProvider{}
		fakeClient.NewAppStreamProviderReturnsOnCall(1, fakeAppStreamProvider2)

		go discovery.Start(ctx)

		Eventually(fakeClient.NewAppStreamProviderCallCount).Should(Equal(2))
		Eventually(fakeAppStreamProvider1.StartCallCount).Should(Equal(1))
		Eventually(fakeAppStreamProvider2.StartCallCount).Should(Equal(1))
		Eventually(fakeAppStreamProvider1.CloseCallCount).Should(Equal(1))

		Consistently(fakeAppStreamProvider1.StartCallCount, 200*time.Millisecond).Should(Equal(1))
		Consistently(fakeAppStreamProvider1.CloseCallCount, 200*time.Millisecond).Should(Equal(1))
		Consistently(fakeAppStreamProvider2.StartCallCount, 200*time.Millisecond).Should(Equal(1))
		Consistently(fakeAppStreamProvider2.CloseCallCount, 200*time.Millisecond).Should(Equal(0))
	})

	// It("updates an AppWatcher when an app changes size", func() {
	// 	fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{
	// 		{Guid: "33333333-3333-3333-3333-333333333333", Instances: 1, Name: "foo", State: "STARTED"},
	// 	}, nil)
	// 	fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{
	// 		{Guid: "33333333-3333-3333-3333-333333333333", Instances: 2, Name: "foo", State: "STARTED"},
	// 	}, nil)

	// 	e := exporter.New(fakeClient, fakeWatcherManager)

	// 	go e.Start(ctx, 100*time.Millisecond)

	// 	Eventually(fakeWatcherManager.UpdateAppInstancesCallCount).Should(Equal(1))

	// 	app := fakeWatcherManager.UpdateAppInstancesArgsForCall(0)
	// 	Expect(app.Guid).To(Equal("33333333-3333-3333-3333-333333333333"))
	// 	Expect(app.Instances).To(Equal(2))
	// })
})
