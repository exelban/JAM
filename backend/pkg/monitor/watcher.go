package monitor

import (
	"context"
	"github.com/exelban/uptime/pkg/dialer"
	"github.com/exelban/uptime/pkg/notify"
	"github.com/exelban/uptime/store"
	"github.com/exelban/uptime/store/engine"
	"github.com/exelban/uptime/types"
	"log"
	"sync"
	"time"
)

type watcher struct {
	dialer  *dialer.Dialer
	notify  *notify.Notify
	history store.Store
	host    types.Host

	status    types.StatusType
	lastCheck time.Time

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// run - runs check loop for host
func (w *watcher) run() {
	if w.history == nil {
		w.history = engine.NewLocal(w.host.History)
	}

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

// check - call to the host and check host status
func (w *watcher) check() {
	resp := w.dialer.Dial(w.ctx, &w.host)
	if !resp.OK {
		return
	}

	w.mu.Lock()
	resp.Status = w.host.Status(resp.Code, resp.Bytes)
	w.lastCheck = time.Now()
	w.history.Add(resp)
	w.validate()
	w.mu.Unlock()

	log.Printf("[DEBUG] %s: %s status (last: %v)", w.host.String(), w.status, resp.Status)
}

// validate - checks success and failure thresholds. And settings the host status
func (w *watcher) validate() {
	checks := w.history.Checks()

	if len(checks) > 0 && len(checks) >= w.host.FailureThreshold && w.status != types.DOWN {
		ok := true
		for _, v := range checks[len(checks)-w.host.FailureThreshold:] {
			if v.Status {
				ok = false
			}
		}

		if ok {
			newStatus := types.DOWN
			if w.status != types.Unknown {
				if err := w.notify.Set(newStatus, w.host.String()); err != nil {
					log.Print(err)
				}
			}
			w.history.SetStatus(newStatus)
			w.status = newStatus
		}
	}

	if len(checks) > 0 && len(checks) >= w.host.SuccessThreshold && w.status != types.UP {
		ok := true
		for _, v := range checks[len(checks)-w.host.SuccessThreshold:] {
			if !v.Status {
				ok = false
			}
		}

		if ok {
			newStatus := types.UP
			if w.status != types.Unknown {
				if err := w.notify.Set(newStatus, w.host.String()); err != nil {
					log.Print(err)
				}
			}
			w.history.SetStatus(newStatus)
			w.status = newStatus
		}
	}

	if w.status == "" {
		w.status = types.Unknown
	}
}
