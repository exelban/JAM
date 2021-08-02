package runner

import (
	"context"
	"github.com/exelban/cheks/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWatcher_check(t *testing.T) {
	ts, status, shutdown := dialServer(0)
	defer shutdown()

	w := &watcher{
		dialer: NewDialer(1),
		host: types.Host{
			URL:              ts.URL,
			SuccessThreshold: 2,
			FailureThreshold: 3,
			Success: &types.Success{
				Code: []int{200},
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
			host: types.Host{},
		}
		w.validate()
		require.Equal(t, types.Unknown, w.status)
	})

	t.Run("1 thresholds", func(t *testing.T) {
		w := &watcher{
			host: types.Host{
				SuccessThreshold: 1,
				FailureThreshold: 1,
			},
		}

		w.validate()
		require.Equal(t, types.Unknown, w.status)

		w.history = append(w.history, s{
			value: false,
		})
		w.history = append(w.history, s{
			value: true,
		})
		w.validate()
		require.Equal(t, types.UP, w.status)

		w.history = append(w.history, s{
			value: false,
		})
		w.validate()
		require.Equal(t, types.DOWN, w.status)
	})

	t.Run("success", func(t *testing.T) {
		w := &watcher{
			host: types.Host{
				SuccessThreshold: 3,
				FailureThreshold: 2,
			},
		}

		for i := 0; i < 6; i++ {
			w.history = append(w.history, s{
				value: false,
			})
			w.validate()
		}
		w.history = append(w.history, s{
			value: true,
		})
		w.validate()
		w.history = append(w.history, s{
			value: true,
		})
		w.validate()
		require.Equal(t, types.DOWN, w.status)

		w.history = append(w.history, s{
			value: true,
		})
		w.validate()
		require.Equal(t, types.UP, w.status)

		w.history = append(w.history, s{
			value: false,
		})
		w.validate()
		require.Equal(t, types.UP, w.status)
		w.history = append(w.history, s{
			value: false,
		})
		w.validate()
		require.Equal(t, types.DOWN, w.status)
	})

	t.Run("failure", func(t *testing.T) {
		w := &watcher{
			host: types.Host{
				SuccessThreshold: 2,
				FailureThreshold: 3,
			},
		}

		for i := 0; i < 6; i++ {
			w.history = append(w.history, s{
				value: true,
			})
			w.validate()
		}
		w.history = append(w.history, s{
			value: false,
		})
		w.validate()
		w.history = append(w.history, s{
			value: false,
		})
		w.validate()
		require.Equal(t, types.UP, w.status)
		w.history = append(w.history, s{
			value: false,
		})
		w.validate()
		require.Equal(t, types.DOWN, w.status)

		w.history = append(w.history, s{
			value: true,
		})
		w.validate()
		require.Equal(t, types.DOWN, w.status)
	})
}
