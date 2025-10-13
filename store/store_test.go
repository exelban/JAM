package store

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/exelban/JAM/types"
	"github.com/stretchr/testify/require"
)

func TestStore_AddResponse(t *testing.T) {
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
				require.NoError(t, s.AddResponse(ctx, "test", &types.HttpResponse{Code: i, Timestamp: now.Add(-time.Minute * time.Duration(i))}))
			}
			h, err := s.FindResponses(ctx, "test")
			require.NoError(t, err)
			require.Equal(t, count, len(h))
		})
	}
}

func TestStore_ResponseHistory(t *testing.T) {
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
					_ = s.AddResponse(ctx, "test", &types.HttpResponse{Code: i, Timestamp: now.Add(-time.Minute * time.Duration(i))})
				}(i)
			}
			wg.Wait()
			history, err := s.FindResponses(ctx, "test")
			require.NoError(t, err)
			require.Equal(t, count, len(history))

			for i, h := range history {
				require.Equal(t, count-i-1, h.Code)
			}

			require.Equal(t, now.Unix(), history[len(history)-1].Timestamp.Unix())
		})
	}
}

func TestStore_DeleteResponse(t *testing.T) {
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
				require.NoError(t, s.AddResponse(ctx, "test", &types.HttpResponse{Code: i, Timestamp: now.Add(-time.Minute * time.Duration(i))}))
			}
			responses, err := s.FindResponses(ctx, "test")
			require.NoError(t, err)
			require.Equal(t, count, len(responses))

			t.Run("delete half of the responses", func(t *testing.T) {
				half := count / 2
				keys := make([]time.Time, 0)
				for i := 0; i < half; i++ {
					keys = append(keys, responses[i].Timestamp)
				}
				require.NoError(t, s.DeleteResponse(ctx, "test", keys))

				responses, err = s.FindResponses(ctx, "test")
				require.NoError(t, err)
				require.Equal(t, count-half, len(responses))
			})

			t.Run("delete all responses", func(t *testing.T) {
				keys := make([]time.Time, 0)
				for i := 0; i < len(responses); i++ {
					keys = append(keys, responses[i].Timestamp)
				}
				require.NoError(t, s.DeleteResponse(ctx, "test", keys))

				responses, err = s.FindResponses(ctx, "test")
				require.NoError(t, err)
				require.Empty(t, responses)
			})
		})
	}
}

func TestStore_Hosts(t *testing.T) {
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
				hostID := fmt.Sprintf("host-%d", i)
				require.NoError(t, s.AddResponse(ctx, hostID, &types.HttpResponse{Timestamp: now.Add(-time.Minute * time.Duration(i))}))
			}
			hosts, err := s.Hosts(ctx)
			require.NoError(t, err)
			require.Equal(t, count, len(hosts))
		})
	}
}

func TestStore_AddEvent(t *testing.T) {
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
				require.NoError(t, s.AddIncident(ctx, "test", &types.Incident{
					StartTS: now.Add(-time.Minute * time.Duration(i)),
					EndTS:   &now,
				}))
				time.Sleep(time.Millisecond)
			}
			e, err := s.FindIncidents(ctx, "test", 0, 0)
			require.NoError(t, err)
			require.Equal(t, count, len(e))

			eventToFinish := e[3]
			require.NoError(t, s.EndIncident(ctx, "test", eventToFinish.ID, now))
			e, err = s.FindIncidents(ctx, "test", -1, -1)
			require.NoError(t, err)
			require.NotNil(t, e[3].EndTS)
			require.Equal(t, now.Unix(), e[3].EndTS.Unix())
		})
	}
}

func TestStore_FindEvents(t *testing.T) {
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
				require.NoError(t, s.AddIncident(ctx, "test", &types.Incident{
					StartTS: now.Add(-time.Minute * time.Duration(i)),
					EndTS:   &now,
				}))
				time.Sleep(time.Millisecond)
			}

			t.Run("no skip and no limit", func(t *testing.T) {
				events, err := s.FindIncidents(ctx, "test", 0, 0)
				require.NoError(t, err)
				require.Equal(t, count, len(events))
				require.Equal(t, count, events[0].ID)
				require.Equal(t, 1, events[len(events)-1].ID)
			})

			t.Run("no skip with limit", func(t *testing.T) {
				events, err := s.FindIncidents(ctx, "test", 0, 10)
				require.NoError(t, err)
				require.Equal(t, 10, len(events))
				require.Equal(t, count, events[0].ID)
				require.Equal(t, count-9, events[len(events)-1].ID)
			})

			t.Run("with skip no limit", func(t *testing.T) {
				events, err := s.FindIncidents(ctx, "test", 5, 0)
				require.NoError(t, err)
				require.Equal(t, count-5, len(events))
				require.Equal(t, count-5, events[0].ID)
				require.Equal(t, 1, events[len(events)-1].ID)
			})

			t.Run("with skip and limit", func(t *testing.T) {
				limitedAndSkipped, err := s.FindIncidents(ctx, "test", 5, 3)
				require.NoError(t, err)
				require.Equal(t, 3, len(limitedAndSkipped))
				require.Equal(t, count-5, limitedAndSkipped[0].ID)
				require.Equal(t, count-7, limitedAndSkipped[len(limitedAndSkipped)-1].ID)
			})

			t.Run("get last event", func(t *testing.T) {
				e, err := s.FindIncidents(ctx, "test", 0, 1)
				require.NoError(t, err)
				require.Len(t, e, 1)
				require.Equal(t, count, e[0].ID)
			})
		})
	}
}

func TestStore_aggregation(t *testing.T) {
	t.Run("one day history", func(t *testing.T) {
		ctx := context.Background()
		s, err := New(ctx, "memory", nil)
		require.NoError(t, err)
		require.NotNil(t, s)

		start := time.Now().Add(-24 * time.Hour).Truncate(time.Hour * 24)
		today := GenerateHistory(s, start, "test")

		require.NoError(t, Aggregate(ctx, s))

		history, err := s.FindResponses(ctx, "test")
		require.NoError(t, err)
		require.Equal(t, today+1, len(history))
	})
	t.Run("random days history back", func(t *testing.T) {
		ctx := context.Background()
		s, err := New(ctx, "memory", nil)
		require.NoError(t, err)
		require.NotNil(t, s)

		days := rand.Intn(1000-500) + 500
		start := time.Now().Add(-24 * time.Hour * time.Duration(days)).Truncate(time.Hour * 24)
		today := GenerateHistory(s, start, "test")

		require.NoError(t, Aggregate(ctx, s))

		history, err := s.FindResponses(ctx, "test")
		require.NoError(t, err)
		require.Equal(t, today+days, len(history))
	})
	t.Run("uptime, status and responseTime type per day", func(t *testing.T) {
		ctx := context.Background()
		s, err := New(ctx, "memory", nil)
		require.NoError(t, err)
		require.NotNil(t, s)

		daysNum := rand.Intn(100-50) + 50
		start := time.Now().Add(-24 * time.Hour * time.Duration(daysNum)).Truncate(time.Hour * 24)
		_ = GenerateHistory(s, start, "test")

		history, err := s.FindResponses(ctx, "test")
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

		history, err = s.FindResponses(ctx, "test")
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
