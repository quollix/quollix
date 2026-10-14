package metrics

import (
	"sync"
	"time"

	"server/configs"

	u "github.com/quollix/common/utils"
)

type MetricsMonitorImpl struct {
	CollectionService MetricsCollectionService
	ConfigsRepository configs.ConfigsRepository
	IntervalConfigKey string
	Enabled           bool
	lifecycleMutex    sync.Mutex
	cycleMutex        sync.Mutex
	stopMonitoring    chan struct{}
}

func NewMetricsMonitor(service MetricsCollectionService, repository configs.ConfigsRepository, intervalConfigKey string, enabled bool) *MetricsMonitorImpl {
	return &MetricsMonitorImpl{CollectionService: service, ConfigsRepository: repository, IntervalConfigKey: intervalConfigKey, Enabled: enabled}
}

func (m *MetricsMonitorImpl) StartMonitoring() {
	m.lifecycleMutex.Lock()
	defer m.lifecycleMutex.Unlock()
	if !m.Enabled || m.stopMonitoring != nil {
		return
	}
	m.stopMonitoring = make(chan struct{})
	go m.run(m.stopMonitoring)
}

func (m *MetricsMonitorImpl) StopMonitoring() {
	m.lifecycleMutex.Lock()
	defer m.lifecycleMutex.Unlock()
	if m.stopMonitoring == nil {
		return
	}
	close(m.stopMonitoring)
	m.stopMonitoring = nil
}

func (m *MetricsMonitorImpl) run(stop <-chan struct{}) {
	interval := time.Minute
	for {
		m.cycleMutex.Lock()
		select {
		case <-stop:
			m.cycleMutex.Unlock()
			return
		default:
		}
		err := m.CollectionService.CollectAndPush()
		m.cycleMutex.Unlock()
		if err != nil {
			u.Logger.Error(err)
		}
		nextInterval, err := m.readInterval()
		if err != nil {
			u.Logger.Error(err)
		} else {
			interval = nextInterval
		}
		select {
		case <-stop:
			return
		case <-time.After(interval):
		}
	}
}

func (m *MetricsMonitorImpl) readInterval() (time.Duration, error) {
	value, err := m.ConfigsRepository.GetConfig(m.IntervalConfigKey)
	if err != nil {
		return 0, err
	}
	interval, err := time.ParseDuration(value + "s")
	if err != nil || interval <= 0 {
		return 0, u.Logger.NewError("Invalid metrics collection interval", "key", m.IntervalConfigKey, "value", value)
	}
	return interval, nil
}

func (m *MetricsMonitorImpl) RunCycle() error {
	m.cycleMutex.Lock()
	defer m.cycleMutex.Unlock()
	return m.CollectionService.CollectAndPush()
}
