package runner

import (
	"context"
	"github.com/exelban/cheks/app/types"
	"log"
	"sync"
	"time"
)

// Monitor - main service which track the hosts liveness
type Monitor struct {
	Config *types.Config
	Dialer *Dialer

	watchers []*watcher
	mu       sync.RWMutex
}

// Run - run the monitor. Creates a jobs for each host in the separate threads
func (m *Monitor) Run(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, host := range m.Config.Hosts {
		ctx_, cancel := context.WithCancel(ctx)

		w := &watcher{
			dialer: m.Dialer,

			host:   host,
			status: types.Unknown,
			ctx:    ctx_,
			cancel: cancel,
		}
		go m.watch(w)

		m.watchers = append(m.watchers, w)
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

// History - returns the status history for all hosts
func (m *Monitor) History() map[string]map[time.Time]bool {
	list := make(map[string]map[time.Time]bool)

	m.mu.RLock()
	for _, w := range m.watchers {
		w.mu.RLock()
		{
			history := make(map[time.Time]bool)
			for _, p := range w.history {
				history[p.time] = p.value
			}
			list[w.host.String()] = history
		}
		w.mu.RUnlock()
	}
	m.mu.RUnlock()

	return list
}

// Services - return the services for the app
func (m *Monitor) Services() map[string]types.Service {
	list := make(map[string]types.Service)

	m.mu.RLock()
	for _, w := range m.watchers {
		w.mu.RLock()
		{
			history := make(map[string]bool)
			for _, p := range w.history {
				history[p.time.Format("15:04:05 02.01.2006")] = p.value
			}
			list[w.host.String()] = types.Service{
				Status:    w.status,
				LastCheck: w.lastCheck.Format("02.01.2006 15:04:05"),
				History:   history,
			}
		}
		w.mu.RUnlock()
	}
	m.mu.RUnlock()

	return list
}

// watch - hosts job with ticker. Runs the liveness check for host
func (m *Monitor) watch(w *watcher) {
	log.Printf("[INFO] %s: new watcher", w.host.String())

	time.Sleep(w.host.InitialDelayInterval)
	w.check()

	ticker := time.NewTicker(w.host.RetryInterval)
	for {
		select {
		case <-ticker.C:
			w.check()
		case <-w.ctx.Done():
			log.Printf("[DEBUG] %s: stopped", w.host.String())
			ticker.Stop()
			return
		}
	}
}
