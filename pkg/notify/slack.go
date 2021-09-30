package notify

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type Slack struct {
	url      string
	username string
	channel  string

	timeout time.Duration
}

func (s *Slack) send(str string) error {
	b, _ := json.Marshal(struct {
		Username string `json:"username,omitempty"`
		Channel  string `json:"channel,omitempty"`
		Text     string `json:"text,omitempty"`
	}{
		Username: s.username,
		Channel:  s.channel,
		Text:     str,
	})

	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

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

	if buf.String() != "ok" {
		return errors.New("non-ok response returned from Slack")
	}

	return nil
}
