package notify

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSlack_send(t *testing.T) {
	router := http.NewServeMux()

	router.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		req := struct {
			Text string `json:"text,omitempty"`
		}{}
		_ = json.Unmarshal(b, &req)

		if req.Text == "timeout" {
			time.Sleep(time.Millisecond * 20)
		} else if req.Text == "error" {
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{\"ok\": true}"))
	})
	ts := httptest.NewServer(router)
	defer func() {
		ts.Close()
	}()

	slack := &Slack{
		url:     ts.URL,
		token:   "test",
		channel: "test",
		timeout: time.Millisecond * 10,
	}

	require.NoError(t, slack.send("test"))
	require.Error(t, slack.send("error"))
	require.Error(t, slack.send("timeout"))
}
