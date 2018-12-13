package app

import (
	"context"
	"log"
	"time"

	"github.com/alphagov/paas-prometheus-exporter/cf"
	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/client_golang/prometheus"
)

// Struct to store all names related to an app (app name, space name, org name) so we can track if these have changed
// for a given app and if so delete and recreate its app watcher
type appMetadata struct {
	appName   string
	spaceName string
	orgName   string
}

func newAppMetadata(app cfclient.App) appMetadata {
	return appMetadata{
		appName:   app.Name,
		spaceName: app.SpaceData.Entity.Name,
		orgName:   app.SpaceData.Entity.OrgData.Entity.Name,
	}
}

type Discovery struct {
	client               cf.Client
	prometheusRegisterer prometheus.Registerer
	checkInterval        time.Duration
	appMetadataByGUID    map[string]appMetadata
	watchers             map[string]*Watcher
}

func NewDiscovery(
	client cf.Client,
	prometheusRegisterer prometheus.Registerer,
	checkInterval time.Duration,
) *Discovery {
	return &Discovery{
		client:               client,
		prometheusRegisterer: prometheusRegisterer,
		checkInterval:        checkInterval,
		appMetadataByGUID:    make(map[string]appMetadata),
		watchers:             make(map[string]*Watcher),
	}
}

func (s *Discovery) Start(ctx context.Context) {
	go s.run(ctx)
}

func (s *Discovery) run(ctx context.Context) {
	timer := time.NewTicker(s.checkInterval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			log.Println("checking for new apps")
			err := s.checkForNewApps()
			if err != nil {
				log.Fatal(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Discovery) checkForNewApps() error {
	apps, err := s.client.ListAppsWithSpaceAndOrg()
	if err != nil {
		return err
	}

	running := map[string]bool{}

	for _, app := range apps {
		if app.State == "STARTED" {
			running[app.Guid] = true

			if appMetadata, ok := s.appMetadataByGUID[app.Guid]; ok {
				if appMetadata != newAppMetadata(app) {
					// Either the name of the app, the name of it's space or the name of it's org has changed
					s.deleteWatcher(app.Guid)
					err := s.createNewWatcher(app)
					if err != nil {
						return err
					}
				} else {
					// notify watcher that instances may have changed
					s.watchers[app.Guid].UpdateAppInstances(app.Instances)
				}
			} else {
				// new app
				err := s.createNewWatcher(app)
				if err != nil {
					return err
				}
			}
		}
	}

	for appGUID, _ := range s.appMetadataByGUID {
		if ok := running[appGUID]; !ok {
			s.deleteWatcher(appGUID)
		}
	}
	return nil
}

func (s *Discovery) createNewWatcher(app cfclient.App) error {
	watcher, err := NewWatcher(app, s.prometheusRegisterer, s.client.NewAppStreamProvider(app.Guid))
	if err != nil {
		return err
	}

	s.watchers[app.Guid] = watcher
	s.appMetadataByGUID[app.Guid] = newAppMetadata(app)

	go watcher.Run()

	return nil
}

func (s *Discovery) deleteWatcher(appGUID string) {
	s.watchers[appGUID].Close()
	delete(s.watchers, appGUID)
	delete(s.appMetadataByGUID, appGUID)
}
