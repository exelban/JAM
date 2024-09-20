package monitor

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/exelban/JAM/pkg/dialer"
	"github.com/exelban/JAM/pkg/notify"
	"github.com/exelban/JAM/store"
	"github.com/exelban/JAM/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestWatcher_check(t *testing.T) {
	ts, status, shutdown := srv(0)
	defer shutdown()
	ctx := context.Background()

	sn := 2
	fn := 3
	ri := 100 * time.Millisecond

	w := &watcher{
		dialer: dialer.New(1),
		notify: &notify.Notify{},
		store:  store.NewMemory(ctx),
		host: &types.Host{
			URL: ts.URL,
			Conditions: &types.Success{
				Code: []int{200},
			},
			SuccessThreshold: &sn,
			FailureThreshold: &fn,
			Interval:         &ri,
		},
		ctx: ctx,
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
	for i := 0; i < 30; i++ {
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
	ctx := context.Background()

	t.Run("no thresholds", func(t *testing.T) {
		w := &watcher{
			host:   &types.Host{},
			notify: &notify.Notify{},
			store:  store.NewMemory(ctx),
		}
		w.validate(true)
		require.Equal(t, types.UP, w.status)
		w.validate(false)
		require.Equal(t, types.DOWN, w.status)
		w.validate(true)
		require.Equal(t, types.UP, w.status)
	})

	t.Run("min thresholds", func(t *testing.T) {
		sn := 2
		fn := 2

		w := &watcher{
			notify: &notify.Notify{},
			store:  store.NewMemory(ctx),
			host: &types.Host{
				ID:               id(),
				SuccessThreshold: &sn,
				FailureThreshold: &fn,
			},
		}

		w.validate(true)
		require.Equal(t, types.Unknown, w.status)

		w.validate(false)
		require.Equal(t, types.Unknown, w.status)
		w.validate(false)
		require.Equal(t, types.DOWN, w.status)
		w.validate(true)
		require.Equal(t, types.DOWN, w.status)
		w.validate(true)
		require.Equal(t, types.UP, w.status)

		w.validate(false)
		require.Equal(t, types.UP, w.status)
	})

	t.Run("success", func(t *testing.T) {
		sn := 3
		fn := 2

		w := &watcher{
			notify: &notify.Notify{},
			store:  store.NewMemory(ctx),
			host: &types.Host{
				ID:               id(),
				SuccessThreshold: &sn,
				FailureThreshold: &fn,
			},
		}

		for i := 0; i < 6; i++ {
			w.validate(false)
		}
		w.validate(true)
		w.validate(true)
		require.Equal(t, types.DOWN, w.status)

		w.validate(true)
		require.Equal(t, types.UP, w.status)

		w.validate(false)
		require.Equal(t, types.UP, w.status)
		w.validate(false)
		require.Equal(t, types.DOWN, w.status)
	})

	t.Run("failure", func(t *testing.T) {
		sn := 2
		fn := 3

		w := &watcher{
			notify: &notify.Notify{},
			store:  store.NewMemory(ctx),
			host: &types.Host{
				ID:               id(),
				SuccessThreshold: &sn,
				FailureThreshold: &fn,
			},
		}

		for i := 0; i < 6; i++ {
			w.validate(true)
		}
		w.validate(false)
		w.validate(false)
		require.Equal(t, types.UP, w.status)
		w.validate(false)
		require.Equal(t, types.DOWN, w.status)

		w.validate(true)
		require.Equal(t, types.DOWN, w.status)
	})
}

func id() string {
	n := 12
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%X", b)
}
