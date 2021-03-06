package cloud

import (
	"context"
	"sync"
	"time"

	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/model/host"
	"github.com/pkg/errors"
)

var globalMockState *mockState

func init() {
	globalMockState = &mockState{
		instances: map[string]MockInstance{},
	}
}

// MockInstance mocks a running server that Evergreen knows about. It contains
// fields that can be set to change the response the cloud manager returns
// when this mock instance is queried for.
type MockInstance struct {
	IsUp               bool
	IsSSHReachable     bool
	Status             CloudStatus
	SSHOptions         []string
	TimeTilNextPayment time.Duration
	DNSName            string
	OnUpRan            bool
}

type MockProvider interface {
	Len() int
	Reset()
	Get(string) MockInstance
	Set(string, MockInstance)
	IterIDs() <-chan string
	IterInstances() <-chan MockInstance
}

func GetMockProvider() MockProvider {
	return globalMockState
}

type mockState struct {
	instances map[string]MockInstance
	mutex     sync.RWMutex
}

func (m *mockState) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.instances = map[string]MockInstance{}
}

func (m *mockState) Len() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return len(m.instances)
}

func (m *mockState) IterIDs() <-chan string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	out := make(chan string, len(m.instances))

	for id := range m.instances {
		out <- id
	}

	close(out)
	return out
}

func (m *mockState) Get(id string) MockInstance {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.instances[id]
}

func (m *mockState) Set(id string, instance MockInstance) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.instances[id] = instance
}

func (m *mockState) IterInstances() <-chan MockInstance {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	out := make(chan MockInstance, len(m.instances))

	for _, node := range m.instances {
		out <- node
	}

	close(out)
	return out
}

// mockManager implements the Manager interface for testing
// purposes. It contains a map of MockInstances that it knows about
// which its various functions return information about. Once set before
// testing, this map should only be touched either through the associated
// cloud manager functions, or in association with the mutex.
type mockManager struct {
	Instances map[string]MockInstance
	mutex     *sync.RWMutex
}

func makeMockManager() Manager {
	return &mockManager{
		Instances: globalMockState.instances,
		mutex:     &globalMockState.mutex,
	}
}

func (mockMgr *mockManager) SpawnHost(ctx context.Context, h *host.Host) (*host.Host, error) {
	l := mockMgr.mutex
	l.Lock()
	defer l.Unlock()
	mockMgr.Instances[h.Id] = MockInstance{
		IsUp:               false,
		IsSSHReachable:     false,
		Status:             StatusInitializing,
		SSHOptions:         []string{},
		TimeTilNextPayment: time.Duration(0),
		DNSName:            "",
	}
	return h, nil
}

// get the status of an instance
func (mockMgr *mockManager) GetInstanceStatus(ctx context.Context, host *host.Host) (CloudStatus, error) {
	l := mockMgr.mutex
	l.RLock()
	instance, ok := mockMgr.Instances[host.Id]
	l.RUnlock()
	if !ok {
		return StatusUnknown, errors.Errorf("unable to fetch host: %s", host.Id)
	}

	return instance.Status, nil
}

// get instance DNS
func (mockMgr *mockManager) GetDNSName(ctx context.Context, host *host.Host) (string, error) {
	l := mockMgr.mutex
	l.RLock()
	instance, ok := mockMgr.Instances[host.Id]
	l.RUnlock()
	if !ok {
		return "", errors.Errorf("unable to fetch host: %s", host.Id)
	}
	return instance.DNSName, nil
}

func (_ *mockManager) GetSettings() ProviderSettings {
	return &mockManager{}
}

func (_ *mockManager) Validate() error {
	return nil
}

// terminate an instance
func (mockMgr *mockManager) TerminateInstance(ctx context.Context, host *host.Host, user string) error {
	l := mockMgr.mutex
	l.Lock()
	defer l.Unlock()
	instance, ok := mockMgr.Instances[host.Id]
	if !ok {
		return errors.Errorf("unable to fetch host: %s", host.Id)
	}
	if host.Status == evergreen.HostTerminated {
		return errors.Errorf("Cannot terminate %s; already marked as terminated!", host.Id)
	}

	instance.Status = StatusTerminated
	mockMgr.Instances[host.Id] = instance

	return errors.WithStack(host.Terminate(user))
}

func (mockMgr *mockManager) Configure(ctx context.Context, settings *evergreen.Settings) error {
	//no-op. maybe will need to load something from settings in the future.
	return nil
}

func (mockMgr *mockManager) IsUp(ctx context.Context, host *host.Host) (bool, error) {
	l := mockMgr.mutex
	l.RLock()
	instance, ok := mockMgr.Instances[host.Id]
	l.RUnlock()
	if !ok {
		return false, errors.Errorf("unable to fetch host: %s", host.Id)
	}
	return instance.IsUp, nil
}

func (mockMgr *mockManager) OnUp(ctx context.Context, host *host.Host) error {
	l := mockMgr.mutex
	l.Lock()
	defer l.Unlock()
	instance, ok := mockMgr.Instances[host.Id]
	if !ok {
		return errors.Errorf("unable to fetch host: %s", host.Id)
	}
	instance.OnUpRan = true
	mockMgr.Instances[host.Id] = instance

	return nil
}

func (mockMgr *mockManager) GetSSHOptions(host *host.Host, keyPath string) ([]string, error) {
	l := mockMgr.mutex
	l.RLock()
	instance, ok := mockMgr.Instances[host.Id]
	l.RUnlock()
	if !ok {
		return []string{}, errors.Errorf("unable to fetch host: %s", host.Id)
	}
	return instance.SSHOptions, nil
}

func (mockMgr *mockManager) TimeTilNextPayment(host *host.Host) time.Duration {
	l := mockMgr.mutex
	l.RLock()
	instance, ok := mockMgr.Instances[host.Id]
	l.RUnlock()
	if !ok {
		return time.Duration(0)
	}
	return instance.TimeTilNextPayment
}

func (mockMgr *mockManager) GetInstanceStatuses(ctx context.Context, hosts []host.Host) ([]CloudStatus, error) {
	if len(hosts) != 2 {
		return nil, errors.New("expecting 2 hosts")
	}
	return []CloudStatus{StatusRunning, StatusRunning}, nil
}

// CostForDuration for the mock returns 1 dollar per minute up
func (m *mockManager) CostForDuration(ctx context.Context, h *host.Host, start, end time.Time, s *evergreen.Settings) (float64, error) {
	return end.Sub(start).Minutes(), nil
}
