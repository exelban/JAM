package runner

import (
	"context"
	"github.com/exelban/cheks/store"
	"github.com/exelban/cheks/store/engine"
	"github.com/exelban/cheks/types"
	"log"
	"sync"
	"time"
)

type watcher struct {
	dialer  *Dialer
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
	w.mu.Lock()
	defer w.mu.Unlock()

	resp := w.dialer.Dial(w.ctx, &w.host)
	if !resp.OK {
		return
	}
	resp.Status = w.host.Status(resp.Code, resp.Bytes)
	w.lastCheck = time.Now()

	w.history.Add(resp)
	w.validate()

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
			w.history.SetStatus(types.DOWN)
			w.status = types.DOWN
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
			w.history.SetStatus(types.UP)
			w.status = types.UP
		}
	}

	if w.status == "" {
		w.status = types.Unknown
	}
}
