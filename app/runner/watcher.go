package runner

import (
	"context"
	"github.com/exelban/cheks/app/types"
	"log"
	"sync"
	"time"
)

type s struct {
	time       time.Time
	value      bool
	statusCode int
	body       []byte
}

type watcher struct {
	dialer *Dialer

	host      types.Host
	status    types.StatusType
	lastCheck time.Time
	history   []s

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

func (w *watcher) check() {
	w.mu.Lock()
	defer w.mu.Unlock()

	responseCode, b, ok := w.dialer.Dial(w.ctx, &w.host)
	if !ok {
		return
	}
	status := w.host.Status(responseCode, b)
	w.lastCheck = time.Now()

	if len(w.history) >= 100 {
		w.history = w.history[1:len(w.history)]
	}
	w.history = append(w.history, s{
		time:       time.Now(),
		value:      status,
		statusCode: responseCode,
		body:       b,
	})

	w.validate()

	log.Printf("[DEBUG] %s: %s status (last: %v)", w.host.String(), w.status, status)
}

// validate - checks success and failure thresholds. And settings the host status
func (w *watcher) validate() {
	if len(w.history) > 0 && len(w.history) >= w.host.FailureThreshold && w.status != types.DOWN {
		ok := true
		for _, v := range w.history[len(w.history)-w.host.FailureThreshold:] {
			if v.value {
				ok = false
			}
		}

		if ok {
			w.status = types.DOWN
		}
	}

	if len(w.history) > 0 && len(w.history) >= w.host.SuccessThreshold && w.status != types.UP {
		ok := true
		for _, v := range w.history[len(w.history)-w.host.SuccessThreshold:] {
			if !v.value {
				ok = false
			}
		}

		if ok {
			w.status = types.UP
		}
	}

	if w.status == "" {
		w.status = types.Unknown
	}
}
