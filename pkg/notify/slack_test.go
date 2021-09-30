package notify

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/pkgz/rest"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSlack_send(t *testing.T) {
	router := chi.NewRouter()

	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		req := struct {
			Text string `json:"text,omitempty"`
		}{}
		_ = json.Unmarshal(b, &req)

		if req.Text == "timeout" {
			time.Sleep(time.Millisecond * 20)
		} else if req.Text == "error" {
			rest.ErrorResponse(w, r, http.StatusInternalServerError, nil, "error")
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	ts := httptest.NewServer(router)
	defer func() {
		ts.Close()
	}()

	slack := &Slack{
		url:      ts.URL,
		username: "test",
		channel:  "test",
		timeout:  time.Millisecond * 10,
	}

	require.NoError(t, slack.send("test"))
	require.Error(t, slack.send("error"))
	require.Error(t, slack.send("timeout"))
}
