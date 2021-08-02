package runner

import (
	"context"
	"github.com/exelban/cheks/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestMonitor_Run(t *testing.T) {
	ts, status, shutdown := dialServer(0)
	defer shutdown()

	m := Monitor{
		Config: &types.Config{
			Hosts: []types.Host{
				{
					Name: "host-0",
					URL:  ts.URL,
					Success: &types.Success{
						Code: []int{200},
					},
					SuccessThreshold:     2,
					FailureThreshold:     3,
					InitialDelayInterval: time.Millisecond * 10,
					RetryInterval:        time.Millisecond * 30,
					TimeoutInterval:      time.Millisecond * 100,
				},
			},
		},
		Dialer: NewDialer(3),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, m.Run(ctx))

	t.Run("must be up", func(t *testing.T) {
		time.Sleep(time.Millisecond * 100)
		require.Equal(t, types.UP, m.Status()["host-0"])
	})

	t.Run("must does down", func(t *testing.T) {
		status.Store(false)
		time.Sleep(time.Millisecond * 100)
		require.Equal(t, types.DOWN, m.Status()["host-0"])
	})

	t.Run("must does down", func(t *testing.T) {
		status.Store(true)
		time.Sleep(time.Millisecond * 30)
		require.Equal(t, types.DOWN, m.Status()["host-0"])
	})

	cancel()
}

func TestMonitor_Status(t *testing.T) {
	m := Monitor{
		watchers: []*watcher{
			{
				host: types.Host{
					Name: "host-0",
				},
			},
			{
				host: types.Host{
					Name: "host-1",
				},
				status: types.UP,
			},
			{
				host: types.Host{
					Name: "host-2",
				},
				status: types.DOWN,
			},
		},
	}

	list := m.Status()
	require.Equal(t, types.Unknown, list["host-0"])
	require.Equal(t, types.UP, list["host-1"])
	require.Equal(t, types.DOWN, list["host-2"])
}

func TestMonitor_History(t *testing.T) {
	m := Monitor{
		watchers: []*watcher{
			{
				host: types.Host{
					Name: "host-0",
				},
				history: []s{
					{
						time:  time.Now(),
						value: false,
					},
				},
			},
			{
				host: types.Host{
					Name: "host-1",
				},
				status: types.UP,
				history: []s{
					{
						time:  time.Now(),
						value: false,
					},
					{
						time:  time.Now(),
						value: true,
					},
				},
			},
		},
	}

	list := m.History()
	require.Len(t, list["host-0"], 1)
	require.Len(t, list["host-1"], 2)
}

func TestMonitor_Services(t *testing.T) {
	t1 := time.Now().Add(-time.Minute)
	t2 := time.Now().Add(time.Minute)

	m := Monitor{
		watchers: []*watcher{
			{
				lastCheck: t1,
				host: types.Host{
					Name: "host-0",
				},
				history: []s{
					{
						time:  time.Now(),
						value: false,
					},
				},
			},
			{
				host: types.Host{
					Name: "host-1",
				},
				status:    types.UP,
				lastCheck: t2,
				history: []s{
					{
						time:  time.Now(),
						value: false,
					},
					{
						time:  time.Now().Add(-time.Minute),
						value: true,
					},
				},
			},
		},
	}

	list := m.Services()

	require.Equal(t, t1.Format("02.01.2006 15:04:05"), list["host-0"].LastCheck)
	require.Equal(t, t2.Format("02.01.2006 15:04:05"), list["host-1"].LastCheck)

	require.Len(t, list["host-0"].History, 1)
	require.Len(t, list["host-1"].History, 2)
}
