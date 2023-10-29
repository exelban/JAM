package notify

import (
	"context"
	"fmt"
	"github.com/exelban/uptime/types"
	"log"
	"strings"
	"sync"
	"time"
)

//go:generate moq -out mock_test.go . notify

type notify interface {
	send(str string) error
}

type Notify struct {
	clients []notify

	mu sync.Mutex
}

func New(ctx context.Context, cfg *types.Cfg) (*Notify, error) {
	n := &Notify{}

	if cfg.Alerts.InitializationMessage == nil {
		t := true
		cfg.Alerts.InitializationMessage = &t
	}

	if cfg.Alerts.Slack != nil {
		slack := &Slack{
			url:     "https://slack.com/api/chat.postMessage",
			token:   cfg.Alerts.Slack.Token,
			channel: cfg.Alerts.Slack.Channel,
			timeout: time.Second * 10,
		}
		n.clients = append(n.clients, slack)
		log.Print("[INFO] Slack notifications enabled")
	}
	if cfg.Alerts.Telegram != nil {
		telegram := &Telegram{
			token:   cfg.Alerts.Telegram.Token,
			chatIDs: cfg.Alerts.Telegram.ChatIDs,
		}
		n.clients = append(n.clients, telegram)
		log.Print("[INFO] Telegram notifications enabled")
	}

	if *cfg.Alerts.InitializationMessage {
		for _, client := range n.clients {
			if err := client.send("I'm online"); err != nil {
				return nil, fmt.Errorf("send up message: %w", err)
			}
		}
	}

	if cfg.LivenessInterval != "" {
		log.Printf("[INFO] Liveness interval is enabled every %s", cfg.LivenessInterval)

		duration, err := time.ParseDuration(cfg.LivenessInterval)
		if err != nil {
			return nil, fmt.Errorf("parse liveness interval: %w", err)
		}
		tk := time.NewTicker(duration)

		go func() {
		loop:
			for {
				select {
				case <-tk.C:
					for _, client := range n.clients {
						if err := client.send("Liveness check"); err != nil {
							log.Printf("[ERROR] Liveness check: %s", err)
						}
					}
				case <-ctx.Done():
					tk.Stop()
					if cfg.Alerts.ShutdownMessage {
						for _, client := range n.clients {
							if err := client.send("Going offline..."); err != nil {
								log.Printf("[ERROR] send shutdown message: %s", err)
							}
						}
					}
					break loop
				}
			}
		}()
	}

	return n, nil
}

func (n *Notify) Set(status types.StatusType, name string) error {
	text := fmt.Sprintf("`%s` has a new status: %s", name, strings.ToUpper(string(status)))

	n.mu.Lock()
	defer n.mu.Unlock()

	for _, c := range n.clients {
		if err := c.send(text); err != nil {
			return err
		}
	}

	return nil
}