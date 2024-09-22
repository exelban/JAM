package store

import (
	"context"
	"github.com/exelban/JAM/types"
	"sort"
	"sync"
	"time"
)

type Memory struct {
	history map[string]map[time.Time]*types.HttpResponse
	sync.RWMutex
}

func NewMemory(ctx context.Context) *Memory {
	return &Memory{
		history: make(map[string]map[time.Time]*types.HttpResponse),
	}
}

// Close closes the store.
func (m *Memory) Close() error {
	return nil
}

// Add adds a new response to the history of the given ID.
func (m *Memory) Add(ctx context.Context, id string, r *types.HttpResponse) error {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.history[id]; !ok {
		m.history[id] = make(map[time.Time]*types.HttpResponse)
	}

	m.history[id][r.Timestamp.UTC()] = r
	return nil
}

// Keys returns the keys of the store.
func (m *Memory) Keys(ctx context.Context) ([]string, error) {
	m.RLock()
	defer m.RUnlock()

	keys := []string{}
	for k := range m.history {
		keys = append(keys, k)
	}

	return keys, nil
}

// History returns the history of the given ID.
func (m *Memory) History(ctx context.Context, id string, limit int) ([]*types.HttpResponse, error) {
	m.RLock()
	defer m.RUnlock()

	if _, ok := m.history[id]; !ok {
		return []*types.HttpResponse{}, nil
	}

	res := make([]*types.HttpResponse, 0, len(m.history[id]))
	for _, r := range m.history[id] {
		res = append(res, r)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Timestamp.Before(res[j].Timestamp)
	})

	if limit > 0 && len(res) > limit {
		res = res[len(res)-limit:]
	}

	return res, nil
}

// Delete deletes the given keys from the history of the given ID.
func (m *Memory) Delete(ctx context.Context, id string, keys []time.Time) error {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.history[id]; !ok {
		return nil
	}

	for _, key := range keys {
		for ts := range m.history[id] {
			if ts == key.UTC() {
				delete(m.history[id], ts)
				break
			}
		}
	}

	return nil
}
