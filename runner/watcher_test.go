package runner

import (
	"context"
	"github.com/exelban/cheks/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWatcher_check(t *testing.T) {
	ts, status, shutdown := dialServer(0)
	defer shutdown()

	w := &watcher{
		dialer: NewDialer(1),
		host: config.Host{
			URL:              ts.URL,
			SuccessThreshold: 2,
			FailureThreshold: 3,
			Success: &config.Success{
				Code: []int{200},
			},
			History: &config.HistoryCounts{
				Check: 100,
			},
		},
		ctx: context.Background(),
	}

	w.check()
	require.Equal(t, config.Unknown, w.status)

	w.check()
	require.Equal(t, config.UP, w.status)

	status.Store(false)
	w.check()
	require.Equal(t, config.UP, w.status)
	w.check()
	require.Equal(t, config.UP, w.status)
	w.check()
	require.Equal(t, config.DOWN, w.status)

	status.Store(true)
	w.check()
	require.Equal(t, config.DOWN, w.status)
	w.check()
	require.Equal(t, config.UP, w.status)

	// reach the history limit
	for i := 0; i < 100; i++ {
		w.check()
	}

	status.Store(false)
	w.check()
	require.Equal(t, config.UP, w.status)
	w.check()
	require.Equal(t, config.UP, w.status)
	w.check()
	require.Equal(t, config.DOWN, w.status)

	status.Store(true)
	w.check()
	require.Equal(t, config.DOWN, w.status)
	w.check()
	require.Equal(t, config.UP, w.status)
}

func TestWatcher_validate(t *testing.T) {
	t.Run("no thresholds", func(t *testing.T) {
		w := &watcher{
			host: config.Host{},
		}
		w.validate()
		require.Equal(t, config.Unknown, w.status)
	})

	t.Run("1 thresholds", func(t *testing.T) {
		w := &watcher{
			host: config.Host{
				SuccessThreshold: 1,
				FailureThreshold: 1,
			},
		}

		w.validate()
		require.Equal(t, config.Unknown, w.status)

		w.checks = append(w.checks, config.HttpResponse{
			Status: false,
		})
		w.checks = append(w.checks, config.HttpResponse{
			Status: true,
		})
		w.validate()
		require.Equal(t, config.UP, w.status)

		w.checks = append(w.checks, config.HttpResponse{
			Status: false,
		})
		w.validate()
		require.Equal(t, config.DOWN, w.status)
	})

	t.Run("success", func(t *testing.T) {
		w := &watcher{
			host: config.Host{
				SuccessThreshold: 3,
				FailureThreshold: 2,
			},
		}

		for i := 0; i < 6; i++ {
			w.checks = append(w.checks, config.HttpResponse{
				Status: false,
			})
			w.validate()
		}
		w.checks = append(w.checks, config.HttpResponse{
			Status: true,
		})
		w.validate()
		w.checks = append(w.checks, config.HttpResponse{
			Status: true,
		})
		w.validate()
		require.Equal(t, config.DOWN, w.status)

		w.checks = append(w.checks, config.HttpResponse{
			Status: true,
		})
		w.validate()
		require.Equal(t, config.UP, w.status)

		w.checks = append(w.checks, config.HttpResponse{
			Status: false,
		})
		w.validate()
		require.Equal(t, config.UP, w.status)
		w.checks = append(w.checks, config.HttpResponse{
			Status: false,
		})
		w.validate()
		require.Equal(t, config.DOWN, w.status)
	})

	t.Run("failure", func(t *testing.T) {
		w := &watcher{
			host: config.Host{
				SuccessThreshold: 2,
				FailureThreshold: 3,
			},
		}

		for i := 0; i < 6; i++ {
			w.checks = append(w.checks, config.HttpResponse{
				Status: true,
			})
			w.validate()
		}
		w.checks = append(w.checks, config.HttpResponse{
			Status: false,
		})
		w.validate()
		w.checks = append(w.checks, config.HttpResponse{
			Status: false,
		})
		w.validate()
		require.Equal(t, config.UP, w.status)
		w.checks = append(w.checks, config.HttpResponse{
			Status: false,
		})
		w.validate()
		require.Equal(t, config.DOWN, w.status)

		w.checks = append(w.checks, config.HttpResponse{
			Status: true,
		})
		w.validate()
		require.Equal(t, config.DOWN, w.status)
	})
}
