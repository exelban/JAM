package api

import (
	"fmt"
	"github.com/exelban/cheks/types"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
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

func server() (*httptest.Server, func()) {
	apiRest := &Rest{
		Monitor: &monitorMock{
			StatusFunc: func() map[string]types.StatusType {
				return make(map[string]types.StatusType)
			},
			ServicesFunc: func() []types.Service {
				return []types.Service{}
			},
		},
	}

	ts := httptest.NewServer(apiRest.Router())
	shutdown := func() {
		ts.Close()
	}

	return ts, shutdown
}
