package runner

import (
	"context"
	"github.com/exelban/cheks/config"
	"log"
	"sync"
	"time"
)

type watcher struct {
	dialer *Dialer
	host   config.Host

	status    config.StatusType
	lastCheck time.Time

	checks  []config.HttpResponse
	success []config.HttpResponse
	failure []config.HttpResponse

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// run - runs check loop for host
func (w *watcher) run() {
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

	if len(w.checks) >= w.host.History.Check {
		w.checks = w.checks[1:len(w.checks)]
	}
	w.checks = append(w.checks, resp)

	w.validate()

	log.Printf("[DEBUG] %s: %s status (last: %v)", w.host.String(), w.status, resp.Status)
}

// validate - checks success and failure thresholds. And settings the host status
func (w *watcher) validate() {
	if len(w.checks) > 0 && len(w.checks) >= w.host.FailureThreshold && w.status != config.DOWN {
		ok := true
		for _, v := range w.checks[len(w.checks)-w.host.FailureThreshold:] {
			if v.Status {
				ok = false
			}
		}

		if ok {
			w.failure = append(w.failure, w.checks[len(w.checks)-1])
			w.status = config.DOWN
		}
	}

	if len(w.checks) > 0 && len(w.checks) >= w.host.SuccessThreshold && w.status != config.UP {
		ok := true
		for _, v := range w.checks[len(w.checks)-w.host.SuccessThreshold:] {
			if !v.Status {
				ok = false
			}
		}

		if ok {
			w.success = append(w.success, w.checks[len(w.checks)-1])
			w.status = config.UP
		}
	}

	if w.status == "" {
		w.status = config.Unknown
	}
}
