package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lithammer/shortuuid/v4"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type Slack struct {
	Channel string `json:"channel" yaml:"channel"`
	Token   string `json:"token" yaml:"token"`
}

type Telegram struct {
	Token   string   `json:"token" yaml:"token"`
	ChatIDs []string `json:"chatIDs" yaml:"chatIDs"`
}

type Alerts struct {
	Slack    *Slack    `json:"slack" yaml:"slack"`
	Telegram *Telegram `json:"telegram" yaml:"telegram"`

	InitializationMessage *bool `json:"initializationMessage" yaml:"initializationMessage"`
	ShutdownMessage       bool  `json:"shutdownMessage" yaml:"shutdownMessage"`
}

type Cfg struct {
	MaxConn int `json:"maxConn" yaml:"maxConn,omitempty"`

	Retry            string `json:"retry" yaml:"retry,omitempty"`
	Timeout          string `json:"timeout" yaml:"timeout,omitempty"`
	InitialDelay     string `json:"initialDelay" yaml:"initialDelay,omitempty"`
	LivenessInterval string `json:"livenessInterval" yaml:"livenessInterval,omitempty"`
	SuccessThreshold int    `json:"successThreshold" yaml:"successThreshold,omitempty"`
	FailureThreshold int    `json:"failureThreshold" yaml:"failureThreshold,omitempty"`

	Success *Success          `json:"success" yaml:"success,omitempty"`
	History *HistoryCounts    `json:"history" yaml:"history,omitempty"`
	Headers map[string]string `json:"headers" yaml:"headers,omitempty"`

	Hosts  []Host `json:"hosts" yaml:"hosts,omitempty"`
	Alerts Alerts `json:"alerts" yaml:"alerts,omitempty"`

	path string    `yaml:"-"`
	FW   chan bool `yaml:"-"`
}

func NewConfig(ctx context.Context, path string) (*Cfg, error) {
	cfg := &Cfg{
		path: path,
		FW:   make(chan bool),
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := cfg.save(); err != nil {
			return nil, fmt.Errorf("save config: %w", err)
		}
	}

	go func() {
		ticker := time.NewTicker(time.Second)
		modTimestamp := time.Time{}
		for {
			select {
			case <-ticker.C:
				fi, err := os.Stat(path)
				if err != nil {
					continue
				}

				if fi.ModTime() != modTimestamp {
					log.Printf("[DEBUG] config changed: %s -> %s",
						modTimestamp.Format(time.RFC3339Nano), fi.ModTime().Format(time.RFC3339Nano))
					cfg.FW <- true
					modTimestamp = fi.ModTime()
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	return cfg, nil
}

// Parse - open a configuration file and parse it into Config based on file type (json or yaml)
func (c *Cfg) Parse() error {
	file, err := os.Open(c.path)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if strings.HasSuffix(c.path, ".yaml") || strings.HasSuffix(c.path, ".yml") {
		log.Print("[DEBUG] detect yaml config")
		return yaml.Unmarshal(bytes, &c)
	} else if strings.HasSuffix(c.path, ".json") {
		log.Print("[DEBUG] detect json config")
		return json.Unmarshal(bytes, &c)
	}

	return fmt.Errorf("unknown configuration format `%s`", c.path)
}

// Validate - trying to guess if host is API or Server, also set default timeout and retry values
func (c *Cfg) Validate() error {
	if c.MaxConn == 0 {
		c.MaxConn = 128
	}
	if c.Retry == "" {
		c.Retry = "60s"
	}
	if c.Timeout == "" {
		c.Timeout = "180s"
	}
	if c.InitialDelay == "" {
		c.InitialDelay = "0"
	}
	if c.Success == nil {
		c.Success = &Success{
			Code: []int{200, 201, 202, 203, 204, 205, 206, 207, 208},
		}
	} else if len(c.Success.Code) == 0 {
		c.Success.Code = []int{200, 201, 202, 203, 204, 205, 206, 207, 208}
	}
	if c.SuccessThreshold == 0 {
		c.SuccessThreshold = 2
	}
	if c.FailureThreshold == 0 {
		c.FailureThreshold = 3
	}

	if c.History == nil {
		c.History = &HistoryCounts{
			Check:   180,
			Success: 30,
			Failure: 30,
		}
	} else if c.History.Check == 0 {
		c.History.Check = 180
	} else if c.History.Success == 0 {
		c.History.Success = 30
	} else if c.History.Failure == 0 {
		c.History.Failure = 30
	}

	for i, host := range c.Hosts {
		c.Hosts[i].ID = shortuuid.New()
		if host.ID != "" {
			c.Hosts[i].ID = host.ID
		} else {
			c.Hosts[i].ID = shortuuid.New()
		}
		c.Hosts[i].Type = host.GetType()

		if host.URL == "" {
			return fmt.Errorf("no url for %s", host.Name)
		}

		if host.Retry == "" {
			c.Hosts[i].Retry = c.Retry
		}
		if host.Timeout == "" {
			c.Hosts[i].Timeout = c.Timeout
		}
		if host.InitialDelay == "" {
			c.Hosts[i].InitialDelay = c.InitialDelay
		}
		if host.Success == nil {
			c.Hosts[i].Success = c.Success
		} else if len(host.Success.Code) == 0 {
			c.Hosts[i].Success.Code = c.Success.Code
		}

		if host.SuccessThreshold == 0 {
			c.Hosts[i].SuccessThreshold = c.SuccessThreshold
		}
		if host.FailureThreshold == 0 {
			c.Hosts[i].FailureThreshold = c.FailureThreshold
		}

		retryInterval, err := time.ParseDuration(c.Hosts[i].Retry)
		if err != nil {
			return fmt.Errorf("retry interval: %w", err)
		}
		c.Hosts[i].RetryInterval = retryInterval

		timeoutInterval, err := time.ParseDuration(c.Hosts[i].Timeout)
		if err != nil {
			return fmt.Errorf("timeout interval: %w", err)
		}
		c.Hosts[i].TimeoutInterval = timeoutInterval

		initialDelayInterval, err := time.ParseDuration(c.Hosts[i].InitialDelay)
		if err != nil {
			return fmt.Errorf("initial delay interval: %w", err)
		}
		c.Hosts[i].InitialDelayInterval = initialDelayInterval

		if host.History == nil {
			c.Hosts[i].History = c.History
		} else if host.History.Check == 0 {
			c.Hosts[i].History.Check = c.History.Check
		} else if host.History.Success == 0 {
			c.Hosts[i].History.Success = c.History.Success
		} else if host.History.Failure == 0 {
			c.Hosts[i].History.Failure = c.History.Failure
		}

		for key, value := range c.Headers {
			if _, ok := c.Hosts[i].Headers[key]; !ok {
				c.Hosts[i].Headers[key] = value
			}
		}

		log.Printf("[DEBUG] ID=%s Name=%s, URL=%s, Type=%s, InitialDelay=%s, Retry=%s, Timeout=%s, SuccessCode=%v, SuccessThreshold=%d, FailureThreshold=%d",
			c.Hosts[i].ID, c.Hosts[i].Name, c.Hosts[i].URL, c.Hosts[i].Type, c.Hosts[i].InitialDelayInterval, c.Hosts[i].RetryInterval,
			c.Hosts[i].TimeoutInterval, c.Hosts[i].Success.Code, c.Hosts[i].SuccessThreshold, c.Hosts[i].FailureThreshold)
	}

	return nil
}

func (c *Cfg) HostsList(ctx context.Context) ([]*Host, error) {
	var hosts []*Host
	for i := range c.Hosts {
		hosts = append(hosts, &c.Hosts[i])
	}
	return hosts, nil
}

func (c *Cfg) FindHost(ctx context.Context, id string) (*Host, error) {
	for _, host := range c.Hosts {
		if host.ID == id {
			return &host, nil
		}
	}
	return nil, fmt.Errorf("host with id %s not found", id)
}

func (c *Cfg) AddHost(ctx context.Context, h *Host) error {
	c.Hosts = append(c.Hosts, *h)

	if c.Validate() != nil {
		return fmt.Errorf("validate config: %w", c.Validate())
	}

	return c.save()
}

func (c *Cfg) DeleteHost(ctx context.Context, id string) error {
	for i, host := range c.Hosts {
		if host.ID == id {
			c.Hosts = append(c.Hosts[:i], c.Hosts[i+1:]...)
			return c.save()
		}
	}
	return fmt.Errorf("host with id %s not found", id)
}

func (c *Cfg) UpdateHost(ctx context.Context, h *Host) error {
	for i, host := range c.Hosts {
		if host.ID == h.ID {
			c.Hosts[i] = *h
			return c.save()
		}
	}
	return fmt.Errorf("host with id %s not found", h.ID)
}

func (c *Cfg) save() error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	if err := os.WriteFile(c.path, b, 0644); err != nil {
		return err
	}

	return nil
}
