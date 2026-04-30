package db

import (
	"github.com/jmoiron/sqlx"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

type DBHealthOptions struct {
	Name                   string
	Enabled                bool
	PingInterval           time.Duration
	FailuresBeforeDegraded int
}

func (o DBHealthOptions) normalize() DBHealthOptions {
	if strings.TrimSpace(o.Name) == "" {
		o.Name = "default"
	}
	if o.PingInterval <= 0 {
		o.PingInterval = 10 * time.Second
	}
	if o.FailuresBeforeDegraded <= 0 {
		o.FailuresBeforeDegraded = 3
	}
	return o
}

type DBHealthState struct {
	Name                string
	DriverName          string
	Enabled             bool
	Status              string
	ConsecutiveFailures int
	LastError           string
	LastErrorAt         time.Time
	LastRecoveredAt     time.Time
	LastUpdatedAt       time.Time
}

type dbHealthMonitor struct {
	db                     *sqlx.DB
	name                   string
	driverName             string
	pingInterval           time.Duration
	failuresBeforeDegraded int
	maxIdleConns           int

	mutex sync.RWMutex
	state DBHealthState
	stop  chan struct{}
}

var dbHealthRegistry = struct {
	mutex    sync.RWMutex
	monitors map[string]*dbHealthMonitor
}{
	monitors: make(map[string]*dbHealthMonitor),
}

func registerDBHealthMonitor(monitor *dbHealthMonitor) {
	dbHealthRegistry.mutex.Lock()
	defer dbHealthRegistry.mutex.Unlock()
	dbHealthRegistry.monitors[monitor.name] = monitor
}

func GetDBHealthState(name string) (DBHealthState, bool) {
	dbHealthRegistry.mutex.RLock()
	monitor, ok := dbHealthRegistry.monitors[name]
	dbHealthRegistry.mutex.RUnlock()
	if !ok {
		return DBHealthState{}, false
	}
	return monitor.snapshot(), true
}

func AllDBHealthStates() []DBHealthState {
	dbHealthRegistry.mutex.RLock()
	defer dbHealthRegistry.mutex.RUnlock()

	names := make([]string, 0, len(dbHealthRegistry.monitors))
	for name := range dbHealthRegistry.monitors {
		names = append(names, name)
	}
	sort.Strings(names)

	states := make([]DBHealthState, 0, len(names))
	for _, name := range names {
		states = append(states, dbHealthRegistry.monitors[name].snapshot())
	}
	return states
}

func newDBHealthMonitor(db *sqlx.DB, driverName string, maxIdleConns int, options DBHealthOptions) *dbHealthMonitor {
	options = options.normalize()

	monitor := &dbHealthMonitor{
		db:                     db,
		name:                   options.Name,
		driverName:             driverName,
		pingInterval:           options.PingInterval,
		failuresBeforeDegraded: options.FailuresBeforeDegraded,
		maxIdleConns:           maxIdleConns,
		stop:                   make(chan struct{}),
		state: DBHealthState{
			Name:       options.Name,
			DriverName: driverName,
			Enabled:    options.Enabled,
			Status:     "healthy",
			// 初始化时已经 ping 成功过，默认健康。
			LastRecoveredAt: time.Now(),
			LastUpdatedAt:   time.Now(),
		},
	}
	registerDBHealthMonitor(monitor)
	return monitor
}

func (m *dbHealthMonitor) start() {
	if !m.state.Enabled {
		return
	}

	go func() {
		ticker := time.NewTicker(m.pingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.probe()
			case <-m.stop:
				return
			}
		}
	}()
}

func (m *dbHealthMonitor) stopLoop() {
	select {
	case <-m.stop:
		return
	default:
		close(m.stop)
	}
}

func (m *dbHealthMonitor) probe() {
	now := time.Now()
	err := m.db.Ping()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.state.LastUpdatedAt = now
	if err == nil {
		if m.state.Status == "degraded" {
			m.state.Status = "healthy"
			m.state.ConsecutiveFailures = 0
			m.state.LastError = ""
			m.state.LastRecoveredAt = now
			log.Printf("db health recovered name=%s driver=%s", m.name, m.driverName)
			return
		}
		m.state.ConsecutiveFailures = 0
		return
	}

	m.state.ConsecutiveFailures++
	m.state.LastError = err.Error()
	m.state.LastErrorAt = now
	m.db.SetMaxIdleConns(0)
	if m.maxIdleConns > 0 {
		m.db.SetMaxIdleConns(m.maxIdleConns)
	}

	if m.state.ConsecutiveFailures < m.failuresBeforeDegraded {
		return
	}
	if m.state.Status == "degraded" {
		return
	}

	m.state.Status = "degraded"
	log.Printf(
		"db health degraded name=%s driver=%s failures=%d err=%s",
		m.name,
		m.driverName,
		m.state.ConsecutiveFailures,
		m.state.LastError,
	)
}

func (m *dbHealthMonitor) snapshot() DBHealthState {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.state
}
