package types

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

// Config - configuration file structure
type Config struct {
	MaxConn int `json:"maxConn" yaml:"maxConn"`

	Retry            string `json:"retry" yaml:"retry"`
	Timeout          string `json:"timeout" yaml:"timeout"`
	InitialDelay     string `json:"initialDelay" yaml:"initialDelay"`
	SuccessCode      []int  `json:"successCode" yaml:"successCode"`
	SuccessThreshold int    `json:"successThreshold" yaml:"successThreshold"`
	FailureThreshold int    `json:"failureThreshold" yaml:"failureThreshold"`

	Hosts []Host `json:"hosts" yaml:"hosts"`
}

var (
	ErrNoHosts = errors.New("no hosts for monitoring")
)

// Parse - open a configuration file and parse it into Config based on file type (json or yaml)
func (c *Config) Parse(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		log.Print("[DEBUG] detect yaml config")
		return yaml.Unmarshal(bytes, &c)
	} else if strings.HasSuffix(path, ".json") {
		log.Print("[DEBUG] detect json config")
		return json.Unmarshal(bytes, &c)
	}

	return fmt.Errorf("unknown configuration format `%s`", path)
}

// Validate - trying to guess if host is API or Server, also set default timeout and retry values
func (c *Config) Validate() error {
	if len(c.Hosts) == 0 {
		return ErrNoHosts
	}

	for i, host := range c.Hosts {
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
		if len(host.SuccessCode) == 0 {
			c.Hosts[i].SuccessCode = c.SuccessCode
		}

		if host.SuccessThreshold == 0 {
			c.Hosts[i].SuccessThreshold = c.SuccessThreshold
		}
		if host.FailureThreshold == 0 {
			c.Hosts[i].FailureThreshold = c.FailureThreshold
		}

		retryInterval, err := time.ParseDuration(c.Hosts[i].Retry)
		if err != nil {
			return errors.Wrap(err, "retry interval")
		}
		c.Hosts[i].RetryInterval = retryInterval

		timeoutInterval, err := time.ParseDuration(c.Hosts[i].Timeout)
		if err != nil {
			return errors.Wrap(err, "timeout interval")
		}
		c.Hosts[i].TimeoutInterval = timeoutInterval

		initialDelayInterval, err := time.ParseDuration(c.Hosts[i].InitialDelay)
		if err != nil {
			return errors.Wrap(err, "initial delay interval")
		}
		c.Hosts[i].InitialDelayInterval = initialDelayInterval

		log.Printf("[INFO] %s settings: InitialDelay=%s, Retry=%s, Timeout=%s, SuccessCode=%v, SuccessThreshold=%d, FailureThreshold=%d",
			c.Hosts[i].String(), c.Hosts[i].InitialDelayInterval, c.Hosts[i].RetryInterval, c.Hosts[i].TimeoutInterval, c.Hosts[i].SuccessCode, c.Hosts[i].SuccessThreshold, c.Hosts[i].FailureThreshold)
	}

	return nil
}
