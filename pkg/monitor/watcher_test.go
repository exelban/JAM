package monitor

import (
	"context"
	"github.com/exelban/cheks/pkg/dialer"
	"github.com/exelban/cheks/pkg/notify"
	"github.com/exelban/cheks/store/engine"
	"github.com/exelban/cheks/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWatcher_check(t *testing.T) {
	ts, status, shutdown := srv(0)
	defer shutdown()

	w := &watcher{
		dialer: dialer.New(1),
		notify: &notify.Notify{},
		history: engine.NewLocal(&types.HistoryCounts{
			Check: 100,
		}),
		host: types.Host{
			URL:              ts.URL,
			SuccessThreshold: 2,
			FailureThreshold: 3,
			Success: &types.Success{
				Code: []int{200},
			},
			History: &types.HistoryCounts{
				Check: 100,
			},
		},
		ctx: context.Background(),
	}

	w.check()
	require.Equal(t, types.Unknown, w.status)

	w.check()
	require.Equal(t, types.UP, w.status)

	status.Store(false)
	w.check()
	require.Equal(t, types.UP, w.status)
	w.check()
	require.Equal(t, types.UP, w.status)
	w.check()
	require.Equal(t, types.DOWN, w.status)

	status.Store(true)
	w.check()
	require.Equal(t, types.DOWN, w.status)
	w.check()
	require.Equal(t, types.UP, w.status)

	// reach the history limit
	for i := 0; i < 100; i++ {
		w.check()
	}

	status.Store(false)
	w.check()
	require.Equal(t, types.UP, w.status)
	w.check()
	require.Equal(t, types.UP, w.status)
	w.check()
	require.Equal(t, types.DOWN, w.status)

	status.Store(true)
	w.check()
	require.Equal(t, types.DOWN, w.status)
	w.check()
	require.Equal(t, types.UP, w.status)
}

func TestWatcher_validate(t *testing.T) {
	t.Run("no thresholds", func(t *testing.T) {
		w := &watcher{
			host:    types.Host{},
			notify:  &notify.Notify{},
			history: engine.NewLocal(&types.HistoryCounts{}),
		}
		w.validate()
		require.Equal(t, types.Unknown, w.status)
	})

	t.Run("1 thresholds", func(t *testing.T) {
		w := &watcher{
			notify: &notify.Notify{},
			host: types.Host{
				SuccessThreshold: 1,
				FailureThreshold: 1,
				History:          &types.HistoryCounts{},
			},
			history: engine.NewLocal(&types.HistoryCounts{
				Check:   10,
				Success: 1,
				Failure: 1,
			}),
		}

		w.validate()
		require.Equal(t, types.Unknown, w.status)

		w.history.Add(types.HttpResponse{
			Status: false,
		})
		w.history.Add(types.HttpResponse{
			Status: true,
		})
		w.validate()
		require.Equal(t, types.UP, w.status)

		w.history.Add(types.HttpResponse{
			Status: false,
		})
		w.validate()
		require.Equal(t, types.DOWN, w.status)
	})

	t.Run("success", func(t *testing.T) {
		w := &watcher{
			notify: &notify.Notify{},
			host: types.Host{
				SuccessThreshold: 3,
				FailureThreshold: 2,
				History:          &types.HistoryCounts{},
			},
			history: engine.NewLocal(&types.HistoryCounts{
				Check:   10,
				Success: 3,
				Failure: 2,
			}),
		}

		for i := 0; i < 6; i++ {
			w.history.Add(types.HttpResponse{
				Status: false,
			})
			w.validate()
		}
		w.history.Add(types.HttpResponse{
			Status: true,
		})
		w.validate()
		w.history.Add(types.HttpResponse{
			Status: true,
		})
		w.validate()
		require.Equal(t, types.DOWN, w.status)

		w.history.Add(types.HttpResponse{
			Status: true,
		})
		w.validate()
		require.Equal(t, types.UP, w.status)

		w.history.Add(types.HttpResponse{
			Status: false,
		})
		w.validate()
		require.Equal(t, types.UP, w.status)
		w.history.Add(types.HttpResponse{
			Status: false,
		})
		w.validate()
		require.Equal(t, types.DOWN, w.status)
	})

	t.Run("failure", func(t *testing.T) {
		w := &watcher{
			notify: &notify.Notify{},
			host: types.Host{
				SuccessThreshold: 2,
				FailureThreshold: 3,
				History:          &types.HistoryCounts{},
			},
			history: engine.NewLocal(&types.HistoryCounts{
				Check:   10,
				Success: 2,
				Failure: 3,
			}),
		}

		for i := 0; i < 6; i++ {
			w.history.Add(types.HttpResponse{
				Status: true,
			})
			w.validate()
		}
		w.history.Add(types.HttpResponse{
			Status: false,
		})
		w.validate()
		w.history.Add(types.HttpResponse{
			Status: false,
		})
		w.validate()
		require.Equal(t, types.UP, w.status)
		w.history.Add(types.HttpResponse{
			Status: false,
		})
		w.validate()
		require.Equal(t, types.DOWN, w.status)

		w.history.Add(types.HttpResponse{
			Status: true,
		})
		w.validate()
		require.Equal(t, types.DOWN, w.status)
	})
}
