package monitor

import (
	"context"
	"github.com/exelban/JAM/store"
	"github.com/exelban/JAM/types"
	"github.com/stretchr/testify/require"
	"math/rand/v2"
	"testing"
	"time"
)

func TestMonitor_Stats(t *testing.T) {
	ctx := context.Background()
	interval := time.Hour
	group := "group"

	t.Run("no hosts", func(t *testing.T) {
		m := Monitor{}
		s, err := m.Stats(ctx)
		require.NoError(t, err)
		require.NotNil(t, s)
		require.False(t, s.IsHost)
		require.Equal(t, types.Unknown, s.Status)
	})
	t.Run("no history", func(t *testing.T) {
		m := Monitor{
			Store: store.NewMemory(ctx),
			watchers: map[string]*watcher{
				"test": {
					host: &types.Host{
						ID:       "test",
						Interval: &interval,
						Group:    &group,
					},
					status: types.UP,
				},
			},
		}

		s, err := m.Stats(ctx)
		require.NoError(t, err)
		require.NotNil(t, s)
		require.False(t, s.IsHost)
		require.Len(t, s.Hosts, 1)
	})

	t.Run("not full history", func(t *testing.T) {
		m := Monitor{
			Store: store.NewMemory(ctx),
			watchers: map[string]*watcher{
				"test": {
					host: &types.Host{
						ID:       "test",
						Interval: &interval,
					},
					status: types.UP,
				},
			},
		}

		for _, r := range generateDays(30) {
			_ = m.Store.AddResponse(ctx, "test", r)
		}
		require.NoError(t, store.Aggregate(ctx, m.Store))

		history, err := m.Store.FindResponses(ctx, "test")
		require.NoError(t, err)
		require.NotEmpty(t, history)

		s, err := m.Stats(ctx)
		require.NoError(t, err)
		require.NotNil(t, s)
		require.False(t, s.IsHost)
		require.Len(t, s.Hosts, 1)

		host := s.Hosts[0]
		require.Len(t, host.Chart.Points, 91)
		require.Len(t, host.Chart.Intervals, 3)

		for _, p := range host.Chart.Points[:60] {
			require.Equal(t, types.Unknown, p.Status)
		}

		for i, p := range host.Chart.Points[60:90] {
			response := history[i]
			require.Equal(t, response.StatusType, p.Status)
			require.Equal(t, p.Timestamp, response.Timestamp.Format("2006-01-02"))
		}
	})

	t.Run("one host", func(t *testing.T) {
		m := Monitor{
			Store: store.NewMemory(ctx),
			watchers: map[string]*watcher{
				"test": {
					host: &types.Host{
						ID:       "test",
						Interval: &interval,
					},
					status: types.UP,
				},
			},
		}

		now := time.Now()
		for i := 0; i < rand.IntN(1000-100)+100; i++ {
			_ = m.Store.AddResponse(ctx, "test", generateResponse(now.Add(-time.Duration(i)*time.Minute)))
		}
		for _, r := range generateDays(rand.IntN(1000-100) + 100) {
			_ = m.Store.AddResponse(ctx, "test", r)
		}
		require.NoError(t, store.Aggregate(ctx, m.Store))

		history, err := m.Store.FindResponses(ctx, "test")
		require.NoError(t, err)
		require.NotEmpty(t, history)
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		days := make([]*types.HttpResponse, 0)
		today := make([]*types.HttpResponse, 0)
		for _, r := range history {
			if r.Timestamp.After(startOfDay) {
				today = append(today, r)
			} else {
				days = append(days, r)
			}
		}
		history = days[len(days)-90:]
		history = append(history, store.AggregateDay(startOfDay, today))
		require.Len(t, history, 91)

		s, err := m.Stats(ctx)
		require.NoError(t, err)
		require.NotNil(t, s)
		require.False(t, s.IsHost)
		require.Len(t, s.Hosts, 1)

		host := s.Hosts[0]
		require.Len(t, host.Chart.Points, len(history))
		require.Len(t, host.Chart.Intervals, 3)

		for i, p := range host.Chart.Points {
			response := history[len(history)-91+i]
			require.Equal(t, response.StatusType, p.Status)
			require.Equal(t, p.Timestamp, response.Timestamp.Format("2006-01-02"))
		}

		require.Equal(t, "90d", host.Chart.Intervals[0])
		require.Equal(t, "60d", host.Chart.Intervals[1])
		require.Equal(t, "30d", host.Chart.Intervals[2])
	})
	t.Run("one group", func(t *testing.T) {
		m := Monitor{
			Store: store.NewMemory(ctx),
			watchers: map[string]*watcher{
				"test": {
					host: &types.Host{
						ID:       "test",
						Interval: &interval,
						Group:    &group,
					},
					status: types.UP,
				},
			},
		}

		now := time.Now()
		for i := 0; i < rand.IntN(1000-100)+100; i++ {
			_ = m.Store.AddResponse(ctx, "test", generateResponse(now.Add(-time.Duration(i)*time.Minute)))
		}
		for _, r := range generateDays(rand.IntN(1000-100) + 100) {
			_ = m.Store.AddResponse(ctx, "test", r)
		}
		require.NoError(t, store.Aggregate(ctx, m.Store))

		history, err := m.Store.FindResponses(ctx, "test")
		require.NoError(t, err)
		require.NotEmpty(t, history)
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		days := make([]*types.HttpResponse, 0)
		today := make([]*types.HttpResponse, 0)
		for _, r := range history {
			if r.Timestamp.After(startOfDay) {
				today = append(today, r)
			} else {
				days = append(days, r)
			}
		}
		history = days[len(days)-90:]
		history = append(history, store.AggregateDay(startOfDay, today))
		require.Len(t, history, 91)

		s, err := m.Stats(ctx)
		require.NoError(t, err)
		require.NotNil(t, s)
		require.False(t, s.IsHost)
		require.Len(t, s.Hosts, 1)

		host := s.Hosts[0]
		require.Len(t, host.Chart.Points, len(history))
		require.Len(t, host.Chart.Intervals, 3)

		for i, p := range host.Chart.Points {
			response := history[len(history)-91+i]
			require.Equal(t, response.StatusType, p.Status)
			require.Equal(t, p.Timestamp, response.Timestamp.Format("2006-01-02"))
		}

		require.Equal(t, "90d", host.Chart.Intervals[0])
		require.Equal(t, "60d", host.Chart.Intervals[1])
		require.Equal(t, "30d", host.Chart.Intervals[2])
	})

	t.Run("order of hosts", func(t *testing.T) {
		storage := store.NewMemory(ctx)

		m := Monitor{
			Store: storage,
			watchers: map[string]*watcher{
				"test-1": {
					host: &types.Host{
						ID:       "test-1",
						Interval: &interval,
						Index:    0,
					},
				},
				"test-2": {
					host: &types.Host{
						ID:       "test-2",
						Interval: &interval,
						Index:    1,
					},
				},
				"test-3": {
					host: &types.Host{
						ID:       "test-3",
						Interval: &interval,
						Index:    2,
					},
				},
			},
		}

		s, err := m.Stats(ctx)
		require.NoError(t, err)
		require.NotNil(t, s)

		require.Equal(t, "test-1", s.Hosts[0].ID)
		require.Equal(t, "test-2", s.Hosts[1].ID)
		require.Equal(t, "test-3", s.Hosts[2].ID)

		m.watchers["test-1"].host.Index = 1
		m.watchers["test-2"].host.Index = 2
		m.watchers["test-3"].host.Index = 0

		s, err = m.Stats(ctx)
		require.NoError(t, err)
		require.NotNil(t, s)

		require.Equal(t, "test-3", s.Hosts[0].ID)
		require.Equal(t, "test-1", s.Hosts[1].ID)
		require.Equal(t, "test-2", s.Hosts[2].ID)
	})
}

func TestMonitor_StatsByID(t *testing.T) {
	ctx := context.Background()
	interval := time.Second

	t.Run("host not found", func(t *testing.T) {
		m := Monitor{}
		s, err := m.StatsByID(ctx, "not found", false)
		require.Nil(t, s)
		require.Error(t, err)
	})
	t.Run("no history", func(t *testing.T) {
		m := Monitor{
			Store: store.NewMemory(ctx),
			watchers: map[string]*watcher{
				"host": {
					host: &types.Host{
						ID:       "host",
						URL:      "host",
						Interval: &interval,
					},
				},
			},
		}

		now := time.Now()
		s, err := m.StatsByID(ctx, "host", false)
		require.NoError(t, err)
		require.NotNil(t, s)

		require.Equal(t, true, s.IsHost)
		require.Equal(t, types.Unknown, s.Status)
		require.Len(t, s.Hosts, 1)

		host := s.Hosts[0]
		require.Equal(t, types.Unknown, host.Status)
		require.NotEmpty(t, host.Chart)
		require.NotEmpty(t, host.Chart.Points)
		require.Len(t, host.Chart.Points, 90)
		require.Equal(t, 0, host.Uptime)
		require.Equal(t, "0s", host.ResponseTime)

		firstPoint := now.Add(-interval * 90)
		lastPoint := now.Add(-interval)
		require.Equal(t, firstPoint.Unix(), host.Chart.Points[0].TS.Unix())
		require.Equal(t, lastPoint.Unix(), host.Chart.Points[len(host.Chart.Points)-1].TS.Unix())

		for _, p := range host.Chart.Points {
			require.Equal(t, types.Unknown, p.Status)
		}

		require.NotEmpty(t, host.Chart.Intervals)
		require.Len(t, host.Chart.Intervals, 3)
		require.Equal(t, "2m", host.Chart.Intervals[0])
		require.Equal(t, "1m", host.Chart.Intervals[1])
		require.Equal(t, "30s", host.Chart.Intervals[2])

		//require.NotEmpty(t, host.Details)
		//require.NotEmpty(t, host.Details.Uptime)
		//require.Equal(t, "0", host.Details.Uptime[0])
		//require.Equal(t, "0", host.Details.Uptime[1])
		//require.Equal(t, "0", host.Details.Uptime[2])
		//require.NotEmpty(t, host.Details.ResponseTime)
		//require.Equal(t, "0ns", host.Details.ResponseTime[0])
		//require.Equal(t, "0ns", host.Details.ResponseTime[1])
		//require.Equal(t, "0ns", host.Details.ResponseTime[2])
	})
	t.Run("not full history", func(t *testing.T) {
		m := Monitor{
			Store: store.NewMemory(ctx),
			watchers: map[string]*watcher{
				"host": {
					host: &types.Host{
						ID:       "host",
						URL:      "host",
						Interval: &interval,
					},
				},
			},
		}

		responses := []*types.HttpResponse{}
		responseTime := time.Duration(0)
		for i := 0; i < 30; i++ {
			r := generateResponse(time.Now().Add(-time.Duration(i) * time.Second))
			_ = m.Store.AddResponse(ctx, "host", r)
			responses = append(responses, r)
			responseTime += r.Time
		}
		responseTime = responseTime / time.Duration(90)
		require.NoError(t, store.Aggregate(ctx, m.Store))

		s, err := m.StatsByID(ctx, "host", false)
		require.NoError(t, err)
		require.NotNil(t, s)

		host := s.Hosts[0]
		require.Equal(t, types.Unknown, host.Status)
		require.NotEmpty(t, host.Chart)
		require.NotEmpty(t, host.Chart.Points)
		require.Len(t, host.Chart.Points, 90)
		require.Equal(t, responseTime.Truncate(time.Millisecond).String(), host.ResponseTime)

		now := time.Now()
		for i, p := range host.Chart.Points[:60] {
			ts := now.Add(-interval*90 + interval*time.Duration(i))
			require.Equal(t, types.Unknown, p.Status)
			require.Equal(t, ts.Format("2006-01-02 15:04:05"), p.Timestamp)
		}

		for i, p := range host.Chart.Points[60:] {
			response := responses[29-i]
			require.Equal(t, response.StatusType, p.Status)
			require.Equal(t, p.Timestamp, response.Timestamp.Format("2006-01-02 15:04:05"))
		}
	})

	t.Run("few hours", func(t *testing.T) {
		m := Monitor{
			Store: store.NewMemory(ctx),
			watchers: map[string]*watcher{
				"host": {
					host: &types.Host{
						ID:       "host",
						URL:      "host",
						Interval: &interval,
					},
				},
			},
		}

		responses := generateHours(10)
		for _, r := range responses {
			_ = m.Store.AddResponse(ctx, "host", r)
		}
		require.NoError(t, store.Aggregate(ctx, m.Store))

		responseTime := time.Duration(0)
		for _, r := range responses[len(responses)-90:] {
			responseTime += r.Time
		}
		responseTime = responseTime / time.Duration(90)

		s, err := m.StatsByID(ctx, "host", false)
		require.NoError(t, err)
		require.NotNil(t, s)

		host := s.Hosts[0]
		require.Equal(t, types.Unknown, host.Status)
		require.NotEmpty(t, host.Chart)
		require.NotEmpty(t, host.Chart.Points)
		require.Len(t, host.Chart.Points, 90)
		require.Equal(t, responseTime.Truncate(time.Millisecond).String(), host.ResponseTime)

		for i, p := range host.Chart.Points {
			response := responses[len(responses)-90+i]
			require.Equal(t, response.StatusType, p.Status)
			require.Equal(t, p.Timestamp, response.Timestamp.Format("2006-01-02 15:04:05"))
		}
	})
	t.Run("few days history only", func(t *testing.T) {
		m := Monitor{
			Store: store.NewMemory(ctx),
			watchers: map[string]*watcher{
				"host": {
					host: &types.Host{
						ID:       "host",
						URL:      "host",
						Interval: &interval,
					},
				},
			},
		}

		raw := generateDays(90)
		for _, r := range raw {
			_ = m.Store.AddResponse(ctx, "host", r)
		}
		require.NoError(t, store.Aggregate(ctx, m.Store))
		history, err := m.Store.FindResponses(ctx, "host")
		require.NoError(t, err)
		require.Equal(t, 90, len(history))

		responseTime := time.Duration(0)
		for _, r := range history[len(history)-90:] {
			responseTime += r.Time
		}
		responseTime = responseTime / time.Duration(90)

		s, err := m.StatsByID(ctx, "host", false)
		require.NoError(t, err)
		require.NotNil(t, s)

		require.Equal(t, true, s.IsHost)
		require.Equal(t, types.Unknown, s.Status)
		require.Len(t, s.Hosts, 1)

		host := s.Hosts[0]
		require.Equal(t, types.Unknown, host.Status)
		require.NotEmpty(t, host.Chart)
		require.NotEmpty(t, host.Chart.Points)
		require.Len(t, host.Chart.Points, 90)
		require.Equal(t, responseTime.Truncate(time.Millisecond).String(), host.ResponseTime)

		for i, p := range host.Chart.Points {
			response := history[len(history)-90+i]
			require.Equal(t, response.StatusType, p.Status)
			require.Equal(t, p.Timestamp, response.Timestamp.Format("2006-01-02"))
		}
	})
	t.Run("mixed history", func(t *testing.T) {
		m := Monitor{
			Store: store.NewMemory(ctx),
			watchers: map[string]*watcher{
				"host": {
					host: &types.Host{
						ID:       "host",
						URL:      "host",
						Interval: &interval,
					},
				},
			},
		}

		now := time.Now()
		for i := 0; i < 45; i++ {
			_ = m.Store.AddResponse(ctx, "host", generateResponse(now.Add(-time.Duration(i)*time.Minute)))
		}
		for _, r := range generateDays(90) {
			_ = m.Store.AddResponse(ctx, "host", r)
		}
		require.NoError(t, store.Aggregate(ctx, m.Store))
		history, err := m.Store.FindResponses(ctx, "host")
		require.NoError(t, err)
		history = history[len(history)-90:]
		require.Equal(t, 90, len(history))

		responseTime := time.Duration(0)
		for _, r := range history {
			responseTime += r.Time
		}
		responseTime = responseTime / time.Duration(90)

		s, err := m.StatsByID(ctx, "host", false)
		require.NoError(t, err)
		require.NotNil(t, s)

		host := s.Hosts[0]
		require.Equal(t, types.Unknown, host.Status)
		require.NotEmpty(t, host.Chart)
		require.NotEmpty(t, host.Chart.Points)
		require.Len(t, host.Chart.Points, 90)
		require.Equal(t, responseTime.Truncate(time.Millisecond).String(), host.ResponseTime)

		for i, p := range host.Chart.Points[:45] {
			response := history[len(history)-90+i]
			require.Equal(t, response.StatusType, p.Status)
			require.Equal(t, p.Timestamp, response.Timestamp.Format("2006-01-02"))
		}
		for i, p := range host.Chart.Points[45:] {
			response := history[len(history)-45+i]
			require.Equal(t, response.StatusType, p.Status)
			require.Equal(t, p.Timestamp, response.Timestamp.Format("2006-01-02 15:04:05"))
		}
	})

	t.Run("day", func(t *testing.T) {
		m := Monitor{
			Store: store.NewMemory(ctx),
			watchers: map[string]*watcher{
				"host": {
					host: &types.Host{
						ID:       "host",
						URL:      "host",
						Interval: &interval,
					},
				},
			},
		}

		now := time.Now()
		for i := 0; i < 45; i++ {
			_ = m.Store.AddResponse(ctx, "host", generateResponse(now.Add(-time.Duration(i)*time.Minute)))
		}
		for _, r := range generateDays(90) {
			_ = m.Store.AddResponse(ctx, "host", r)
		}
		require.NoError(t, store.Aggregate(ctx, m.Store))
		history, err := m.Store.FindResponses(ctx, "host")
		require.NoError(t, err)
		history = history[len(history)-90:]
		require.Equal(t, 90, len(history))

		responseTime := time.Duration(0)
		for _, r := range history {
			responseTime += r.Time
		}
		responseTime = responseTime / time.Duration(90)

		s, err := m.StatsByID(ctx, "host", true)
		require.NoError(t, err)
		require.NotNil(t, s)

		host := s.Hosts[0]
		require.Equal(t, types.Unknown, host.Status)
		require.NotEmpty(t, host.Chart)
		require.NotEmpty(t, host.Chart.Points)
		//require.Len(t, host.Chart.Points, 90)
		//require.Equal(t, responseTime.Truncate(time.Millisecond).String(), host.ResponseTime)

		require.Equal(t, now.Add(-time.Hour*24*90).Format("2006-01-02"), host.Chart.Points[0].Timestamp)
		require.Equal(t, now.Format("2006-01-02"), host.Chart.Points[len(host.Chart.Points)-1].Timestamp)

		//for i, p := range host.Chart.Points[:45] {
		//	response := history[len(history)-90+i]
		//	require.Equal(t, response.StatusType, p.Status)
		//	require.Equal(t, p.Timestamp, response.Timestamp.Format("2006-01-02"))
		//}
		//for i, p := range host.Chart.Points[45:] {
		//	response := history[len(history)-45+i]
		//	require.Equal(t, response.StatusType, p.Status)
		//	require.Equal(t, p.Timestamp, response.Timestamp.Format("2006-01-02 15:04:05"))
		//}
	})
}

func generateDays(days int) []*types.HttpResponse {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day()-days, 0, 0, 0, 0, now.Location())
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	count := rand.IntN(10000-3000) + 3000
	interval := now.Sub(start) / time.Duration(count)
	res := []*types.HttpResponse{}

	for i := 0; i < count; i++ {
		r := generateResponse(start.Add(interval * time.Duration(i)))
		if r.Timestamp.After(day) {
			continue
		}
		res = append(res, r)
	}

	return res
}
func generateHours(hours int) []*types.HttpResponse {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()-hours, 0, 0, 0, now.Location())
	count := rand.IntN(1000-300) + 300
	interval := now.Sub(start) / time.Duration(count)
	res := []*types.HttpResponse{}

	for i := 0; i < count; i++ {
		res = append(res, generateResponse(start.Add(interval*time.Duration(i))))
	}

	return res
}

func generateResponse(ts time.Time) *types.HttpResponse {
	var status types.StatusType
	switch randInt(1, 3) {
	case 1:
		status = types.UP
	case 2:
		status = types.DOWN
	default:
		status = types.Unknown
	}

	return &types.HttpResponse{
		Timestamp:  ts,
		StatusType: status,
		Time:       time.Duration(randInt(1, 100)) * time.Millisecond,
	}
}

func randInt(min, max int) int {
	return rand.IntN(max-min) + min
}
