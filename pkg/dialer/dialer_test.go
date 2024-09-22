package dialer

import (
	"context"
	"github.com/exelban/JAM/types"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewDialer(t *testing.T) {
	dialer := New(3)
	require.Equal(t, 3, cap(dialer.sem))
}

func TestDialer_Dial(t *testing.T) {
	dialer := New(3)

	ts, _, shutdown := srv(time.Millisecond * 10)
	defer shutdown()
	ctx := context.Background()

	t.Run("wrong method", func(t *testing.T) {
		resp := dialer.Dial(ctx, &types.Host{
			Method: "?",
		})
		require.False(t, resp.OK)
		require.Equal(t, 0, resp.Code)
		require.Empty(t, resp.Bytes)
	})

	t.Run("wrong url", func(t *testing.T) {
		resp := dialer.Dial(ctx, &types.Host{})
		require.False(t, resp.OK)
		require.Equal(t, 0, resp.Code)
		require.Empty(t, resp.Bytes)
	})

	t.Run("semaphore check", func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(9)
		start := time.Now()

		for i := 0; i < 9; i++ {
			go func() {
				resp := dialer.Dial(ctx, &types.Host{
					Method: "GET",
					URL:    ts.URL,
				})
				require.True(t, resp.OK)
				require.Equal(t, http.StatusOK, resp.Code)
				wg.Done()
			}()
		}

		wg.Wait()
		require.Less(t, time.Now().Sub(start).Milliseconds(), int64(50))
		require.Greater(t, time.Now().Sub(start).Milliseconds(), int64(30))
	})

	t.Run("check timeout", func(t *testing.T) {
		timeout := time.Millisecond * 5
		resp := dialer.Dial(ctx, &types.Host{
			Method:          "GET",
			URL:             ts.URL,
			TimeoutInterval: &timeout,
		})
		require.False(t, resp.OK)
		require.Equal(t, 0, resp.Code)
		require.Empty(t, resp.Bytes)
	})
}

func srv(timeout time.Duration) (*httptest.Server, *atomic.Value, func()) {
	router := http.NewServeMux()
	status := atomic.Value{}
	status.Store(true)

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(timeout)
		if status.Load() == true {
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "error", http.StatusInternalServerError)
		}
	})

	ts := httptest.NewServer(router)
	shutdown := func() {
		ts.Close()
	}

	return ts, &status, shutdown
}
