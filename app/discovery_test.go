package app_test

import (
	"context"
	"time"

	dto "github.com/prometheus/client_model/go"

	"github.com/alphagov/paas-prometheus-exporter/app"
	"github.com/alphagov/paas-prometheus-exporter/cf/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/cloudfoundry-community/go-cfclient"
)

const guid = "33333333-3333-3333-3333-333333333333"

var appFixture = cfclient.App{
	Guid:      guid,
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
}

func getMetrics(registry *prometheus.Registry) []*dto.Metric {
	metrics := make([]*dto.Metric, 0)
	metricsFamilies, _ := registry.Gather()
	for _, metricsFamily := range metricsFamilies {
		metrics = append(metrics, metricsFamily.Metric...)
	}
	return metrics
}

func metricHasLabels(metric *dto.Metric, labels map[string]string) bool {
	actualLabels := make(map[string]string)
	for _, pair := range metric.Label {
		actualLabels[*pair.Name] = *pair.Value
	}

	for k, v := range labels {
		if actualValue, ok := actualLabels[k]; !ok || actualValue != v {
			return false
		}
	}

	return true
}

func findMetric(registry *prometheus.Registry, labels map[string]string) *dto.Metric {
	for _, metric := range getMetrics(registry) {
		if metricHasLabels(metric, labels) {
			return metric
		}
	}

	return nil
}

var _ = FDescribe("CheckForNewApps", func() {

	var discovery *app.Discovery
	var fakeClient *mocks.FakeClient
	var ctx context.Context
	var cancel context.CancelFunc
	var registry *prometheus.Registry
	var fakeAppStreamProvider *mocks.FakeAppStreamProvider

	BeforeEach(func() {
		fakeClient = &mocks.FakeClient{}
		fakeAppStreamProvider = &mocks.FakeAppStreamProvider{}
		fakeClient.NewAppStreamProviderReturns(fakeAppStreamProvider)
		registry = prometheus.NewRegistry()
		discovery = app.NewDiscovery(fakeClient, registry, 100*time.Millisecond)
		ctx, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		cancel()
	})

	It("checks for new apps regularly", func() {
		go discovery.Start(ctx)

		Eventually(fakeClient.ListAppsWithSpaceAndOrgCallCount).Should(Equal(2))
	})

	It("creates a new app", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{appFixture}, nil)

		go discovery.Start(ctx)

		Eventually(fakeClient.NewAppStreamProviderCallCount).Should(Equal(1))

		Eventually(func() *dto.Metric {
			return findMetric(registry, map[string]string{
				"guid": guid,
			})
		}).ShouldNot(BeNil())
	})

	It("does not create a new appWatcher if the app state is stopped", func() {
		stoppedApp := appFixture
		stoppedApp.State = "STOPPED"
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{stoppedApp}, nil)

		go discovery.Start(ctx)

		Consistently(fakeClient.NewAppStreamProviderCallCount, 200*time.Millisecond).Should(Equal(0))

		Consistently(func() *dto.Metric {
			return findMetric(registry, map[string]string{
				"guid": guid,
			})
		}, 200*time.Millisecond).Should(BeNil())
	})

	It("deletes an AppWatcher when an app is deleted", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{appFixture}, nil)
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{}, nil)

		go discovery.Start(ctx)

		Eventually(fakeClient.NewAppStreamProviderCallCount).Should(Equal(1))
		Eventually(func() []*dto.Metric { return getMetrics(registry) }).ShouldNot(BeEmpty())

		Eventually(func() *dto.Metric {
			return findMetric(registry, map[string]string{
				"guid": guid,
			})
		}).Should(BeNil())
	})

	It("deletes an AppWatcher when an app is stopped", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{appFixture}, nil)

		stoppedApp := appFixture
		stoppedApp.State = "STOPPED"
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{stoppedApp}, nil)

		go discovery.Start(ctx)

		Eventually(fakeClient.NewAppStreamProviderCallCount).Should(Equal(1))
		Eventually(func() []*dto.Metric { return getMetrics(registry) }).ShouldNot(BeEmpty())

		Eventually(func() *dto.Metric {
			return findMetric(registry, map[string]string{
				"guid": guid,
			})
		}).Should(BeNil())
	})

	It("deletes and recreates an AppWatcher when an app is renamed", func() {
		app1 := appFixture
		app1.Name = "foo"
		fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{app1}, nil)

		app2 := appFixture
		app2.Name = "bar"
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{app2}, nil)

		fakeAppStreamProvider1 := &mocks.FakeAppStreamProvider{}
		fakeClient.NewAppStreamProviderReturnsOnCall(0, fakeAppStreamProvider1)
		fakeAppStreamProvider2 := &mocks.FakeAppStreamProvider{}
		fakeClient.NewAppStreamProviderReturnsOnCall(1, fakeAppStreamProvider2)

		go discovery.Start(ctx)

		Eventually(fakeClient.NewAppStreamProviderCallCount).Should(Equal(2))

		Eventually(func() *dto.Metric {
			return findMetric(registry, map[string]string{
				"guid": guid,
				"app":  "bar",
			})
		}).ShouldNot(BeNil())

		Eventually(func() *dto.Metric {
			return findMetric(registry, map[string]string{
				"guid": guid,
				"app":  "foo",
			})
		}).Should(BeNil())
	})

	It("deletes and recreates an AppWatcher when an app's space is renamed", func() {
		app1 := appFixture
		app1.SpaceData.Entity.Name = "spacename"
		fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{app1}, nil)

		app2 := appFixture
		app2.SpaceData.Entity.Name = "spacenamenew"
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{app2}, nil)

		fakeAppStreamProvider1 := &mocks.FakeAppStreamProvider{}
		fakeClient.NewAppStreamProviderReturnsOnCall(0, fakeAppStreamProvider1)
		fakeAppStreamProvider2 := &mocks.FakeAppStreamProvider{}
		fakeClient.NewAppStreamProviderReturnsOnCall(1, fakeAppStreamProvider2)

		go discovery.Start(ctx)

		Eventually(fakeClient.NewAppStreamProviderCallCount).Should(Equal(2))

		Eventually(func() *dto.Metric {
			return findMetric(registry, map[string]string{
				"guid":  guid,
				"space": "spacenamenew",
			})
		}).ShouldNot(BeNil())

		Eventually(func() *dto.Metric {
			return findMetric(registry, map[string]string{
				"guid":  guid,
				"space": "spacename",
			})
		}).Should(BeNil())
	})

	It("deletes and recreates an AppWatcher when an app's org is renamed", func() {
		app1 := appFixture
		app1.SpaceData.Entity.OrgData.Entity.Name = "orgname"
		fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{app1}, nil)

		app2 := appFixture
		app2.SpaceData.Entity.OrgData.Entity.Name = "orgnamenew"
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{app2}, nil)

		fakeAppStreamProvider1 := &mocks.FakeAppStreamProvider{}
		fakeClient.NewAppStreamProviderReturnsOnCall(0, fakeAppStreamProvider1)
		fakeAppStreamProvider2 := &mocks.FakeAppStreamProvider{}
		fakeClient.NewAppStreamProviderReturnsOnCall(1, fakeAppStreamProvider2)

		go discovery.Start(ctx)

		Eventually(fakeClient.NewAppStreamProviderCallCount).Should(Equal(2))

		Eventually(func() *dto.Metric {
			return findMetric(registry, map[string]string{
				"guid":         guid,
				"organisation": "orgnamenew",
			})
		}).ShouldNot(BeNil())

		Eventually(func() *dto.Metric {
			return findMetric(registry, map[string]string{
				"guid":         guid,
				"organisation": "orgname",
			})
		}).Should(BeNil())
	})

	It("updates an AppWatcher when an app changes size", func() {
		fakeClient.ListAppsWithSpaceAndOrgReturnsOnCall(0, []cfclient.App{appFixture}, nil)

		appWithTwoInstances := appFixture
		appWithTwoInstances.Instances = 2
		fakeClient.ListAppsWithSpaceAndOrgReturns([]cfclient.App{appWithTwoInstances}, nil)

		go discovery.Start(ctx)

		Eventually(fakeClient.NewAppStreamProviderCallCount).Should(Equal(1))

		Eventually(func() *dto.Metric {
			return findMetric(registry, map[string]string{
				"guid":     guid,
				"instance": "0",
			})
		}).ShouldNot(BeNil())

		Eventually(func() *dto.Metric {
			return findMetric(registry, map[string]string{
				"guid":     guid,
				"instance": "1",
			})
		}).ShouldNot(BeNil())
	})
})
