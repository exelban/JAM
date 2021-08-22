package runner

import (
	"context"
	"github.com/exelban/cheks/config"
	"sync"
)

// Monitor - main service which track the hosts liveness
type Monitor struct {
	dialer     *Dialer
	watchers   []*watcher
	tagsColors map[string]string

	mu  sync.RWMutex
	ctx context.Context
}

// Run - run the monitor. Creates a jobs for each host in the separate threads
func (m *Monitor) Run(ctx context.Context, cfg *config.Cfg) error {
	m.mu.Lock()
	{
		if m.ctx != nil {
			m.ctx.Done()
		}
		if m.tagsColors == nil {
			m.tagsColors = make(map[string]string)
		}

		m.ctx = context.Background()
		m.dialer = NewDialer(cfg.MaxConn)
	}
	m.mu.Unlock()

	// add hosts which are does not have watchers
	for _, host := range cfg.Hosts {
		ok := false
		for _, w := range m.watchers {
			if host.Name == w.host.Name && host.URL == w.host.URL {
				ok = true
			}
		}
		if !ok {
			m.add(host)
		}
	}

	// remove watchers which does not present in the config
	for i := len(m.watchers) - 1; i >= 0; i-- {
		ok := false
		for _, host := range cfg.Hosts {
			if host.Name == m.watchers[i].host.Name && host.URL == m.watchers[i].host.URL {
				ok = true
			}
		}
		if !ok {
			m.watchers[i].cancel()
			m.watchers = append(m.watchers[:i], m.watchers[i+1:]...)
		}
	}

	return nil
}

// Status - returns the actual statuses of all hosts
func (m *Monitor) Status() map[string]config.StatusType {
	list := make(map[string]config.StatusType)

	m.mu.RLock()
	for _, w := range m.watchers {
		w.mu.RLock()
		w.validate()
		list[w.host.String()] = w.status
		w.mu.RUnlock()
	}
	m.mu.RUnlock()

	return list
}

// Services - return the services for the app
func (m *Monitor) Services() []config.Service {
	list := []config.Service{}

	m.mu.RLock()
	for _, w := range m.watchers {
		w.mu.RLock()
		{
			var tags []config.Tag
			for _, tag := range w.host.Tags {
				tags = append(tags, config.Tag{
					Name:  tag,
					Color: m.tagsColors[tag],
				})
			}

			list = append(list, config.Service{
				Name: w.host.String(),
				Status: config.Status{
					Value:     w.status,
					Timestamp: w.lastCheck,
				},
				Checks:  w.checks,
				Success: w.success,
				Failure: w.failure,
				Tags:    tags,
			})
		}
		w.mu.RUnlock()
	}
	m.mu.RUnlock()

	return list
}

func (m *Monitor) add(host config.Host) {
	ctx_, cancel := context.WithCancel(m.ctx)

	w := &watcher{
		dialer: m.dialer,
		host:   host,
		ctx:    ctx_,
		cancel: cancel,
	}
	go w.run()

	m.mu.Lock()
	{
		m.watchers = append(m.watchers, w)
		for _, tag := range host.Tags {
			if _, ok := m.tagsColors[tag]; !ok {
				m.tagsColors[tag] = config.RandomColor()
			}
		}
	}
	m.mu.Unlock()

	return
}
