package engine

import (
	"github.com/exelban/cheks/types"
	"sync"
)

type Local struct {
	limits *types.HistoryCounts

	checks  []types.HttpResponse
	success []types.HttpResponse
	failure []types.HttpResponse

	mu sync.RWMutex
}

func NewLocal(limits *types.HistoryCounts) *Local {
	return &Local{
		limits: limits,
	}
}

// Checks - returns the list of checks
func (h *Local) Checks() []types.HttpResponse {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.checks
}

// Success - returns the list of checks
func (h *Local) Success() []types.HttpResponse {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.success
}

// Failure - returns the list of checks
func (h *Local) Failure() []types.HttpResponse {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.failure
}

// Add - add a response to the checks history
func (h *Local) Add(r types.HttpResponse) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.limits.Check == 0 {
		return
	}

	if len(h.checks) >= h.limits.Check {
		h.checks = h.checks[1:]
	}
	h.checks = append(h.checks, r)
}

// SetStatus - set the watcher status and add to the success/failure history
func (h *Local) SetStatus(value types.StatusType) {
	h.mu.Lock()
	defer h.mu.Unlock()

	response := h.checks[len(h.checks)-1]

	switch value {
	case types.UP:
		if h.limits.Success == 0 {
			return
		}

		if len(h.success) >= h.limits.Success {
			h.success = h.success[1:]
		}
		h.success = append(h.success, response)
	case types.DOWN:
		if h.limits.Failure == 0 {
			return
		}

		if len(h.failure) >= h.limits.Failure {
			h.failure = h.failure[1:]
		}
		h.failure = append(h.failure, response)
	}
}
