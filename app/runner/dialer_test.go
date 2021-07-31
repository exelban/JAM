package runner

import (
	"context"
	"github.com/exelban/cheks/app/types"
	"github.com/go-chi/chi/v5"
	"github.com/pkgz/rest"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewDialer(t *testing.T) {
	dialer := NewDialer(3)
	require.Equal(t, 3, cap(dialer.sem))
}

func TestDialer_Dial(t *testing.T) {
	dialer := NewDialer(3)

	ts, _, shutdown := dialServer(time.Millisecond * 10)
	defer shutdown()
	ctx := context.Background()

	t.Run("wrong method", func(t *testing.T) {
		code, ok := dialer.Dial(ctx, &types.Host{
			Method: "?",
		})
		require.False(t, ok)
		require.Equal(t, 0, code)
	})

	t.Run("wrong url", func(t *testing.T) {
		code, ok := dialer.Dial(ctx, &types.Host{})
		require.False(t, ok)
		require.Equal(t, 0, code)
	})

	t.Run("semaphore check", func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(9)
		start := time.Now()

		for i := 0; i < 9; i++ {
			go func() {
				code, ok := dialer.Dial(ctx, &types.Host{
					Method: "GET",
					URL:    ts.URL,
				})
				require.True(t, ok)
				require.Equal(t, http.StatusOK, code)
				wg.Done()
			}()
		}

		wg.Wait()
		require.Less(t, time.Now().Sub(start), time.Millisecond*50)
		require.Greater(t, time.Now().Sub(start), time.Millisecond*30)
	})

	t.Run("check timeout", func(t *testing.T) {
		code, ok := dialer.Dial(ctx, &types.Host{
			Method:          "GET",
			URL:             ts.URL,
			TimeoutInterval: time.Millisecond * 5,
		})
		require.False(t, ok)
		require.Equal(t, 0, code)
	})
}

func dialServer(timeout time.Duration) (*httptest.Server, *atomic.Value, func()) {
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
