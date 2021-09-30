package notify

import (
	"fmt"
	"github.com/exelban/cheks/types"
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

func New(cfg *types.Cfg) (*Notify, error) {
	n := &Notify{}

	if cfg.Alerts.Slack != nil {
		slack := &Slack{
			url:      cfg.Alerts.Slack.URL,
			username: cfg.Alerts.Slack.Username,
			channel:  cfg.Alerts.Slack.Channel,
			timeout:  time.Second * 5,
		}
		n.clients = append(n.clients, slack)
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
