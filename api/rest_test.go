package api

import (
	"fmt"
	"github.com/exelban/cheks/types"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestRest_dashboard(t *testing.T) {
	ts, shutdown := server()
	defer shutdown()
	uri := fmt.Sprintf("%s", ts.URL)

	t.Run("no index.html", func(t *testing.T) {
		resp, err := http.Get(uri)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("wrong list key", func(t *testing.T) {
		file, _ := ioutil.TempFile("/tmp", "*.html")
		defer func() {
			_ = os.Remove(file.Name())
		}()
		_, err := file.Write([]byte(`{{ range $key, $value := .List2 }}<p>{{$key}}</p>{{ end }}`))
		require.NoError(t, err)

		indexPath = file.Name()
		resp, err := http.Get(uri)
		require.NoError(t, err)
		require.NotNil(t, resp)

		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("ok", func(t *testing.T) {
		file, _ := ioutil.TempFile("/tmp", "*.html")
		defer func() {
			_ = os.Remove(file.Name())
		}()
		_, err := file.Write([]byte(`{{ range $key, $value := .List }}<p>{{$key}}</p>{{ end }}`))
		require.NoError(t, err)

		indexPath = file.Name()
		resp, err := http.Get(uri)
		require.NoError(t, err)
		require.NotNil(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)
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
			ServicesFunc: func() map[string]types.Service {
				return make(map[string]types.Service)
			},
		},
		Live: true,
	}

	ts := httptest.NewServer(apiRest.Router())
	shutdown := func() {
		ts.Close()
	}

	return ts, shutdown
}
