package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"time"
)

type Telegram struct {
	token   string
	chatIDs []string

	timeout time.Duration
}

func (t *Telegram) string() string {
	return "telegram"
}

func (t *Telegram) send(str string) error {
	g, _ := errgroup.WithContext(context.Background())
	for _, chatID := range t.chatIDs {
		id := chatID
		g.Go(func() error {
			return t.sendToChat(id, str)
		})
	}
	return g.Wait()
}

func (t *Telegram) sendToChat(chatID, msg string) error {
	b, err := json.Marshal(struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}{
		ChatID: chatID,
		Text:   msg,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("failed to create a new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Timeout: t.timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send a command to the device: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	response := struct {
		Code        int    `json:"code"`
		Description string `json:"description"`
	}{}
	if er := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("unmarshal body to the error: %w", er)
	}

	return errors.New(response.Description)
}
