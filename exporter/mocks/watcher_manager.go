// Code generated by counterfeiter. DO NOT EDIT.
package mocks

import (
	sync "sync"

	events "github.com/alphagov/paas-prometheus-exporter/events"
	exporter "github.com/alphagov/paas-prometheus-exporter/exporter"
	cfclient "github.com/cloudfoundry-community/go-cfclient"
	prometheus "github.com/prometheus/client_golang/prometheus"
)

type FakeWatcherManager struct {
	CreateWatcherStub        func(cfclient.App, prometheus.Registerer) *events.AppWatcher
	createWatcherMutex       sync.RWMutex
	createWatcherArgsForCall []struct {
		arg1 cfclient.App
		arg2 prometheus.Registerer
	}
	createWatcherReturns struct {
		result1 *events.AppWatcher
	}
	createWatcherReturnsOnCall map[int]struct {
		result1 *events.AppWatcher
	}
	DeleteWatcherStub        func(string)
	deleteWatcherMutex       sync.RWMutex
	deleteWatcherArgsForCall []struct {
		arg1 string
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeWatcherManager) CreateWatcher(arg1 cfclient.App, arg2 prometheus.Registerer) *events.AppWatcher {
	fake.createWatcherMutex.Lock()
	ret, specificReturn := fake.createWatcherReturnsOnCall[len(fake.createWatcherArgsForCall)]
	fake.createWatcherArgsForCall = append(fake.createWatcherArgsForCall, struct {
		arg1 cfclient.App
		arg2 prometheus.Registerer
	}{arg1, arg2})
	fake.recordInvocation("CreateWatcher", []interface{}{arg1, arg2})
	fake.createWatcherMutex.Unlock()
	if fake.CreateWatcherStub != nil {
		return fake.CreateWatcherStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.createWatcherReturns
	return fakeReturns.result1
}

func (fake *FakeWatcherManager) CreateWatcherCallCount() int {
	fake.createWatcherMutex.RLock()
	defer fake.createWatcherMutex.RUnlock()
	return len(fake.createWatcherArgsForCall)
}

func (fake *FakeWatcherManager) CreateWatcherCalls(stub func(cfclient.App, prometheus.Registerer) *events.AppWatcher) {
	fake.createWatcherMutex.Lock()
	defer fake.createWatcherMutex.Unlock()
	fake.CreateWatcherStub = stub
}

func (fake *FakeWatcherManager) CreateWatcherArgsForCall(i int) (cfclient.App, prometheus.Registerer) {
	fake.createWatcherMutex.RLock()
	defer fake.createWatcherMutex.RUnlock()
	argsForCall := fake.createWatcherArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeWatcherManager) CreateWatcherReturns(result1 *events.AppWatcher) {
	fake.createWatcherMutex.Lock()
	defer fake.createWatcherMutex.Unlock()
	fake.CreateWatcherStub = nil
	fake.createWatcherReturns = struct {
		result1 *events.AppWatcher
	}{result1}
}

func (fake *FakeWatcherManager) CreateWatcherReturnsOnCall(i int, result1 *events.AppWatcher) {
	fake.createWatcherMutex.Lock()
	defer fake.createWatcherMutex.Unlock()
	fake.CreateWatcherStub = nil
	if fake.createWatcherReturnsOnCall == nil {
		fake.createWatcherReturnsOnCall = make(map[int]struct {
			result1 *events.AppWatcher
		})
	}
	fake.createWatcherReturnsOnCall[i] = struct {
		result1 *events.AppWatcher
	}{result1}
}

func (fake *FakeWatcherManager) DeleteWatcher(arg1 string) {
	fake.deleteWatcherMutex.Lock()
	fake.deleteWatcherArgsForCall = append(fake.deleteWatcherArgsForCall, struct {
		arg1 string
	}{arg1})
	fake.recordInvocation("DeleteWatcher", []interface{}{arg1})
	fake.deleteWatcherMutex.Unlock()
	if fake.DeleteWatcherStub != nil {
		fake.DeleteWatcherStub(arg1)
	}
}

func (fake *FakeWatcherManager) DeleteWatcherCallCount() int {
	fake.deleteWatcherMutex.RLock()
	defer fake.deleteWatcherMutex.RUnlock()
	return len(fake.deleteWatcherArgsForCall)
}

func (fake *FakeWatcherManager) DeleteWatcherCalls(stub func(string)) {
	fake.deleteWatcherMutex.Lock()
	defer fake.deleteWatcherMutex.Unlock()
	fake.DeleteWatcherStub = stub
}

func (fake *FakeWatcherManager) DeleteWatcherArgsForCall(i int) string {
	fake.deleteWatcherMutex.RLock()
	defer fake.deleteWatcherMutex.RUnlock()
	argsForCall := fake.deleteWatcherArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeWatcherManager) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.createWatcherMutex.RLock()
	defer fake.createWatcherMutex.RUnlock()
	fake.deleteWatcherMutex.RLock()
	defer fake.deleteWatcherMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeWatcherManager) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ exporter.WatcherManager = new(FakeWatcherManager)