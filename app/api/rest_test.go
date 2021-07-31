package api

import (
	"fmt"
	"github.com/exelban/cheks/app/types"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRest_Status(t *testing.T) {
	ts, shutdown := server()
	defer shutdown()
	uri := fmt.Sprintf("%s/status", ts.URL)

	t.Run("no data", func(t *testing.T) {
		resp, err := http.Get(uri)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}

func TestRest_History(t *testing.T) {
	ts, shutdown := server()
	defer shutdown()
	uri := fmt.Sprintf("%s/history", ts.URL)

	t.Run("no data", func(t *testing.T) {
		resp, err := http.Get(uri)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}

func server() (*httptest.Server, func()) {
	apiRest := &Rest{
		Monitor: &monitorMock{
			HistoryFunc: func() map[string]map[time.Time]bool {
				return make(map[string]map[time.Time]bool)
			},
			StatusFunc: func() map[string]types.StatusType {
				return make(map[string]types.StatusType)
			},
		},
	}

	ts := httptest.NewServer(apiRest.Router())
	shutdown := func() {
		ts.Close()
	}

	return ts, shutdown
}
