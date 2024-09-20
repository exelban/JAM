package store

import (
	"context"
	"github.com/exelban/JAM/types"
	"github.com/stretchr/testify/require"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
)

func TestStore_Add(t *testing.T) {
	ctx := context.Background()
	list := map[string]func() Interface{
		"memory": func() Interface {
			return NewMemory(ctx)
		},
		"bolt": func() Interface {
			file, err := os.CreateTemp("", "test.db")
			require.NoError(t, err)
			defer os.RemoveAll(file.Name())

			b, err := NewBolt(ctx, file.Name())
			require.NoError(t, err)
			require.NotNil(t, b)

			return b
		},
	}
	now := time.Now()

	for name, f := range list {
		t.Run(name, func(t *testing.T) {
			s := f()
			count := rand.Intn(100-30) + 30
			for i := 0; i < count; i++ {
				require.NoError(t, s.Add(ctx, "test", &types.HttpResponse{Code: i, Timestamp: now.Add(-time.Minute * time.Duration(i))}))
			}
			h, err := s.History(ctx, "test", -1)
			require.NoError(t, err)
			require.Equal(t, count, len(h))
		})
	}
}

func TestStore_Keys(t *testing.T) {
	// TODO
}

func TestStore_History(t *testing.T) {
	ctx := context.Background()
	list := map[string]func() Interface{
		"memory": func() Interface {
			return NewMemory(ctx)
		},
		"bolt": func() Interface {
			file, err := os.CreateTemp("", "test.db")
			require.NoError(t, err)
			defer os.RemoveAll(file.Name())

			b, err := NewBolt(ctx, file.Name())
			require.NoError(t, err)
			require.NotNil(t, b)

			return b
		},
	}

	for name, f := range list {
		t.Run(name, func(t *testing.T) {
			s := f()
			count := rand.Intn(500-100) + 100
			now := time.Now()
			wg := sync.WaitGroup{}
			wg.Add(count)
			for i := 0; i < count; i++ {
				go func(i int) {
					defer wg.Done()
					_ = s.Add(ctx, "test", &types.HttpResponse{Code: i, Timestamp: now.Add(-time.Minute * time.Duration(i))})
				}(i)
			}
			wg.Wait()
			history, err := s.History(ctx, "test", -1)
			require.NoError(t, err)
			require.Equal(t, count, len(history))

			for i, h := range history {
				require.Equal(t, count-i-1, h.Code)
			}

			require.Equal(t, now.Unix(), history[len(history)-1].Timestamp.Unix())
		})
	}
}

func TestStore_Delete(t *testing.T) {
	// TODO
}

func TestStore_aggregation(t *testing.T) {
	t.Run("one day history", func(t *testing.T) {
		ctx := context.Background()
		s, err := New(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, s)

		start := time.Now().Add(-24 * time.Hour).Truncate(time.Hour * 24)
		today := GenerateHistory(s, start, "test")

		require.NoError(t, Aggregate(ctx, s))

		history, err := s.History(ctx, "test", -1)
		require.NoError(t, err)
		require.Equal(t, today+1, len(history))
	})
	t.Run("random days history back", func(t *testing.T) {
		ctx := context.Background()
		s, err := New(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, s)

		days := rand.Intn(1000-500) + 500
		start := time.Now().Add(-24 * time.Hour * time.Duration(days)).Truncate(time.Hour * 24)
		today := GenerateHistory(s, start, "test")

		require.NoError(t, Aggregate(ctx, s))

		history, err := s.History(ctx, "test", -1)
		require.NoError(t, err)
		require.Equal(t, today+days, len(history))
	})
	t.Run("uptime, status and responseTime type per day", func(t *testing.T) {
		ctx := context.Background()
		s, err := New(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, s)

		daysNum := rand.Intn(100-50) + 50
		start := time.Now().Add(-24 * time.Hour * time.Duration(daysNum)).Truncate(time.Hour * 24)
		_ = GenerateHistory(s, start, "test")

		history, err := s.History(ctx, "test", -1)
		require.NoError(t, err)

		days := make(map[time.Time][]*types.HttpResponse)
		for _, h := range history {
			key := time.Date(h.Timestamp.Year(), h.Timestamp.Month(), h.Timestamp.Day(), 0, 0, 0, 0, h.Timestamp.Location())
			days[key] = append(days[key], h)
		}
		require.Equal(t, daysNum+1, len(days))

		type stats struct {
			responseTime time.Duration
			up           int
			count        int
			uptime       float64
		}
		statistics := make(map[time.Time]stats)
		for ts, res := range days {
			stat := stats{
				responseTime: 0,
				up:           0,
				count:        len(res),
			}
			for _, r := range res {
				stat.responseTime += r.Time
				if r.StatusType == types.UP {
					stat.up++
				}
			}
			stat.responseTime = stat.responseTime / time.Duration(len(res))
			stat.uptime = float64(stat.up) / float64(len(res))
			statistics[ts] = stat
		}

		require.NoError(t, Aggregate(ctx, s))

		history, err = s.History(ctx, "test", -1)
		require.NoError(t, err)

		for _, h := range history {
			if h.Uptime == 0 {
				continue
			}
			stat, ok := statistics[h.Timestamp]
			require.True(t, ok)
			require.Equal(t, stat.responseTime, h.Time)
			require.Equal(t, stat.uptime, h.Uptime)
			require.Equal(t, stat.count, h.Count)
			if h.Uptime > 0.95 {
				require.Equal(t, types.UP, h.StatusType)
			} else if h.Uptime > 0.5 {
				require.Equal(t, types.DEGRADED, h.StatusType)
			} else {
				require.Equal(t, types.DOWN, h.StatusType)
			}
		}
	})
}
