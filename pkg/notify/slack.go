package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Slack struct {
	url     string
	token   string
	channel string

	timeout time.Duration
}

func (s *Slack) string() string {
	return "slack"
}

func (s *Slack) send(str string) error {
	b, _ := json.Marshal(struct {
		Username string `json:"username,omitempty"`
		Channel  string `json:"channel,omitempty"`
		Text     string `json:"text,omitempty"`
	}{
		Channel: s.channel,
		Text:    str,
	})

	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.token))

	client := &http.Client{
		Timeout: s.timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	if !strings.Contains(buf.String(), "\"ok\":true") && !strings.Contains(buf.String(), "\"ok\": true") {
		return fmt.Errorf("non-ok (%d) response from Slack: %s", resp.StatusCode, buf.String())
	}

	return nil
}
