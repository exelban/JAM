package monitor

import (
	"github.com/exelban/cheks/store/engine"
	"github.com/exelban/cheks/types"
	"github.com/go-chi/chi/v5"
	"github.com/pkgz/rest"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestMonitor_Run(t *testing.T) {
	ts, status, shutdown := srv(0)
	defer shutdown()

	m := Monitor{}

	hosts := &types.Cfg{
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
				History: &types.HistoryCounts{
					Check:   100,
					Success: 0,
					Failure: 0,
				},
			},
		},
		MaxConn: 3,
	}

	require.NoError(t, m.Run(hosts))

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

	t.Run("add new host", func(t *testing.T) {
		require.Len(t, m.watchers, 1)
		hosts.Hosts = append(hosts.Hosts, types.Host{
			Name: "host-1",
			URL:  ts.URL,
			Success: &types.Success{
				Code: []int{200},
			},
			SuccessThreshold:     2,
			FailureThreshold:     3,
			InitialDelayInterval: time.Millisecond * 10,
			RetryInterval:        time.Millisecond * 30,
			TimeoutInterval:      time.Millisecond * 100,
			History: &types.HistoryCounts{
				Check:   100,
				Success: 0,
				Failure: 0,
			},
		})
		require.NoError(t, m.Run(hosts))
		require.Len(t, m.watchers, 2)
	})

	t.Run("add new host (the same, must not be added)", func(t *testing.T) {
		require.Len(t, m.watchers, 2)
		hosts.Hosts = append(hosts.Hosts, types.Host{
			Name: "host-1",
			URL:  ts.URL,
			Success: &types.Success{
				Code: []int{200},
			},
			SuccessThreshold:     2,
			FailureThreshold:     3,
			InitialDelayInterval: time.Millisecond * 10,
			RetryInterval:        time.Millisecond * 30,
			TimeoutInterval:      time.Millisecond * 100,
			History: &types.HistoryCounts{
				Check:   100,
				Success: 0,
				Failure: 0,
			},
		})
		require.NoError(t, m.Run(hosts))
		require.Len(t, m.watchers, 2)
	})

	t.Run("remove host", func(t *testing.T) {
		require.Len(t, m.watchers, 2)
		hosts.Hosts = hosts.Hosts[:1]
		require.NoError(t, m.Run(hosts))
		require.Len(t, m.watchers, 1)
	})
}

func TestMonitor_Status(t *testing.T) {
	m := Monitor{
		watchers: []*watcher{
			{
				host: types.Host{
					Name: "host-0",
				},
				history: engine.NewLocal(&types.HistoryCounts{}),
			},
			{
				host: types.Host{
					Name: "host-1",
				},
				history: engine.NewLocal(&types.HistoryCounts{}),
				status:  types.UP,
			},
			{
				host: types.Host{
					Name: "host-2",
				},
				history: engine.NewLocal(&types.HistoryCounts{}),
				status:  types.DOWN,
			},
		},
	}

	list := m.Status()
	require.Equal(t, types.Unknown, list["host-0"])
	require.Equal(t, types.UP, list["host-1"])
	require.Equal(t, types.DOWN, list["host-2"])
}

func TestMonitor_Services(t *testing.T) {
	t1 := time.Now().Add(-time.Minute)
	t2 := time.Now().Add(time.Minute)

	h1 := engine.NewLocal(&types.HistoryCounts{
		Check:   10,
		Success: 10,
		Failure: 10,
	})
	h1.Add(types.HttpResponse{
		Timestamp: time.Now(),
		Status:    false,
	})

	h2 := engine.NewLocal(&types.HistoryCounts{
		Check:   10,
		Success: 10,
		Failure: 10,
	})
	h2.Add(types.HttpResponse{
		Timestamp: time.Now(),
		Status:    false,
	})
	h2.Add(types.HttpResponse{
		Timestamp: time.Now().Add(-time.Minute),
		Status:    true,
	})

	m := Monitor{
		watchers: []*watcher{
			{
				lastCheck: t1,
				host: types.Host{
					Name: "b",
				},
				history: h1,
			},
			{
				host: types.Host{
					Name: "a",
				},
				status:    types.UP,
				lastCheck: t2,
				history:   h2,
			},
		},
	}

	list := m.Services()

	require.Equal(t, "b", list[0].Name)
	require.Equal(t, "a", list[1].Name)

	require.Equal(t, t1, list[0].Status.Timestamp)
	require.Equal(t, t2, list[1].Status.Timestamp)
}

func srv(timeout time.Duration) (*httptest.Server, *atomic.Value, func()) {
	router := chi.NewRouter()
	status := atomic.Value{}
	status.Store(true)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(timeout)
		if status.Load() == true {
			rest.OkResponse(w)
		} else {
			rest.ErrorResponse(w, r, http.StatusInternalServerError, nil, "error")
		}
	})

	ts := httptest.NewServer(router)
	shutdown := func() {
		ts.Close()
	}

	return ts, &status, shutdown
}
