package notify

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/exelban/JAM/types"
)

//go:generate moq -out mock_test.go . notify

type notify interface {
	string() string
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
	if cfg.Alerts.SMTP != nil {
		smtp := &SMTP{
			Host:     cfg.Alerts.SMTP.Host,
			Port:     cfg.Alerts.SMTP.Port,
			Username: cfg.Alerts.SMTP.Username,
			Password: cfg.Alerts.SMTP.Password,
			From:     cfg.Alerts.SMTP.From,
			To:       cfg.Alerts.SMTP.To,
		}
		n.clients = append(n.clients, smtp)
		log.Print("[INFO] SMTP notifications enabled")
	}

	if *cfg.Alerts.InitializationMessage {
		for _, client := range n.clients {
			if err := client.send("I'm online"); err != nil {
				log.Printf("[ERROR] send initialization message: %s", err)
			}
		}
	}

	go func() {
	loop:
		for {
			select {
			case <-ctx.Done():
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

	return n, nil
}

func (n *Notify) Set(clients []string, status types.StatusType, name string) error {
	icon := "❌"
	if status == types.UP {
		icon = "✅"
	}
	text := fmt.Sprintf("%s: `%s` has a new status: %s", icon, name, strings.ToUpper(string(status)))

	n.mu.Lock()
	defer n.mu.Unlock()

	for _, c := range n.clients {
		send := false
		if clients != nil && len(clients) > 0 {
			for _, client := range clients {
				if c.string() == client {
					send = true
					break
				}
			}
		} else {
			send = true
		}

		if !send {
			continue
		}

		if err := c.send(text); err != nil {
			return err
		}
	}

	return nil
}
