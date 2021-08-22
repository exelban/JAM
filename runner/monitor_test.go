package runner

import (
	"context"
	"github.com/exelban/cheks/config"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestMonitor_Run(t *testing.T) {
	ts, status, shutdown := dialServer(0)
	defer shutdown()

	m := Monitor{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hosts := &config.Cfg{
		Hosts: []config.Host{
			{
				Name: "host-0",
				URL:  ts.URL,
				Success: &config.Success{
					Code: []int{200},
				},
				SuccessThreshold:     2,
				FailureThreshold:     3,
				InitialDelayInterval: time.Millisecond * 10,
				RetryInterval:        time.Millisecond * 30,
				TimeoutInterval:      time.Millisecond * 100,
				History: &config.HistoryCounts{
					Check:   100,
					Success: 0,
					Failure: 0,
				},
			},
		},
		MaxConn: 3,
	}

	require.NoError(t, m.Run(ctx, hosts))

	t.Run("must be up", func(t *testing.T) {
		time.Sleep(time.Millisecond * 100)
		require.Equal(t, config.UP, m.Status()["host-0"])
	})

	t.Run("must does down", func(t *testing.T) {
		status.Store(false)
		time.Sleep(time.Millisecond * 100)
		require.Equal(t, config.DOWN, m.Status()["host-0"])
	})

	t.Run("must does down", func(t *testing.T) {
		status.Store(true)
		time.Sleep(time.Millisecond * 30)
		require.Equal(t, config.DOWN, m.Status()["host-0"])
	})

	t.Run("add new host", func(t *testing.T) {
		require.Len(t, m.watchers, 1)
		hosts.Hosts = append(hosts.Hosts, config.Host{
			Name: "host-1",
			URL:  ts.URL,
			Success: &config.Success{
				Code: []int{200},
			},
			SuccessThreshold:     2,
			FailureThreshold:     3,
			InitialDelayInterval: time.Millisecond * 10,
			RetryInterval:        time.Millisecond * 30,
			TimeoutInterval:      time.Millisecond * 100,
			History: &config.HistoryCounts{
				Check:   100,
				Success: 0,
				Failure: 0,
			},
		})
		require.NoError(t, m.Run(ctx, hosts))
		require.Len(t, m.watchers, 2)
	})

	t.Run("add new host (the same, must not be added)", func(t *testing.T) {
		require.Len(t, m.watchers, 2)
		hosts.Hosts = append(hosts.Hosts, config.Host{
			Name: "host-1",
			URL:  ts.URL,
			Success: &config.Success{
				Code: []int{200},
			},
			SuccessThreshold:     2,
			FailureThreshold:     3,
			InitialDelayInterval: time.Millisecond * 10,
			RetryInterval:        time.Millisecond * 30,
			TimeoutInterval:      time.Millisecond * 100,
			History: &config.HistoryCounts{
				Check:   100,
				Success: 0,
				Failure: 0,
			},
		})
		require.NoError(t, m.Run(ctx, hosts))
		require.Len(t, m.watchers, 2)
	})

	t.Run("remove host", func(t *testing.T) {
		require.Len(t, m.watchers, 2)
		hosts.Hosts = hosts.Hosts[:1]
		require.NoError(t, m.Run(ctx, hosts))
		require.Len(t, m.watchers, 1)
	})

	cancel()
}

func TestMonitor_Status(t *testing.T) {
	m := Monitor{
		watchers: []*watcher{
			{
				host: config.Host{
					Name: "host-0",
				},
			},
			{
				host: config.Host{
					Name: "host-1",
				},
				status: config.UP,
			},
			{
				host: config.Host{
					Name: "host-2",
				},
				status: config.DOWN,
			},
		},
	}

	list := m.Status()
	require.Equal(t, config.Unknown, list["host-0"])
	require.Equal(t, config.UP, list["host-1"])
	require.Equal(t, config.DOWN, list["host-2"])
}

func TestMonitor_Services(t *testing.T) {
	t1 := time.Now().Add(-time.Minute)
	t2 := time.Now().Add(time.Minute)

	m := Monitor{
		watchers: []*watcher{
			{
				lastCheck: t1,
				host: config.Host{
					Name: "b",
				},
				checks: []config.HttpResponse{
					{
						Timestamp: time.Now(),
						Status:    false,
					},
				},
			},
			{
				host: config.Host{
					Name: "a",
				},
				status:    config.UP,
				lastCheck: t2,
				checks: []config.HttpResponse{
					{
						Timestamp: time.Now(),
						Status:    false,
					},
					{
						Timestamp: time.Now().Add(-time.Minute),
						Status:    true,
					},
				},
			},
		},
	}

	list := m.Services()

	require.Equal(t, "b", list[0].Name)
	require.Equal(t, "a", list[1].Name)

	require.Equal(t, t1, list[0].Status.Timestamp)
	require.Equal(t, t2, list[1].Status.Timestamp)

	require.Len(t, list[0].Checks, 1)
	require.Len(t, list[1].Checks, 2)
}
