package monitor

import (
	"context"
	"github.com/exelban/cheks/pkg/dialer"
	"github.com/exelban/cheks/pkg/notify"
	"github.com/exelban/cheks/store"
	"github.com/exelban/cheks/types"
	"log"
	"sync"
	"time"
)

// Monitor - main service which track the hosts liveness
type Monitor struct {
	dialer *dialer.Dialer
	notify *notify.Notify

	watchers   []*watcher
	tagsColors map[string]string

	mu  sync.RWMutex
	ctx context.Context
}

// Run - run the monitor. Creates a jobs for each host in the separate threads
func (m *Monitor) Run(cfg *types.Cfg) error {
	m.mu.Lock()
	{
		if m.ctx != nil {
			m.ctx.Done()
		}
		if m.tagsColors == nil {
			m.tagsColors = make(map[string]string)
		}

		m.ctx = context.Background()
		m.dialer = dialer.New(cfg.MaxConn)
		n, err := notify.New(m.ctx, cfg)
		if err != nil {
			m.mu.Unlock()
			return err
		}
		m.notify = n
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
			if err := m.add(host); err != nil {
				return err
			}
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
func (m *Monitor) Status() map[string]types.StatusType {
	list := make(map[string]types.StatusType)

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
func (m *Monitor) Services() []types.Service {
	list := []types.Service{}

	start := time.Now()

	m.mu.RLock()
	for _, w := range m.watchers {
		w.mu.RLock()
		{
			var tags []types.Tag
			for _, tag := range w.host.Tags {
				tags = append(tags, types.Tag{
					Name:  tag,
					Color: m.tagsColors[tag],
				})
			}

			list = append(list, types.Service{
				Name: w.host.String(),
				Status: types.Status{
					Value:     w.status,
					Timestamp: w.lastCheck,
				},
				Tags:    tags,
				Checks:  w.history.Checks(),
				Success: w.history.Success(),
				Failure: w.history.Failure(),
			})
		}
		w.mu.RUnlock()
	}
	m.mu.RUnlock()

	log.Printf("[INFO] services list: %v", time.Since(start))

	return list
}

// add - create a watcher for host
func (m *Monitor) add(host types.Host) error {
	history, err := store.New(host.History)
	if err != nil {
		return err
	}

	ctx_, cancel := context.WithCancel(m.ctx)

	w := &watcher{
		dialer:  m.dialer,
		notify:  m.notify,
		history: history,
		host:    host,
		ctx:     ctx_,
		cancel:  cancel,
	}
	go w.run()

	m.mu.Lock()
	{
		m.watchers = append(m.watchers, w)
		for _, tag := range host.Tags {
			if _, ok := m.tagsColors[tag]; !ok {
				m.tagsColors[tag] = types.RandomColor()
			}
		}
	}
	m.mu.Unlock()

	return nil
}
