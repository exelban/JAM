package types

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

type Success struct {
	Code []int   `json:"code" yaml:"code"`
	Body *string `json:"body" yaml:"body"`
}

type HistoryCounts struct {
	Persistent bool `json:"persistent" yaml:"persistent"`

	Check   int `json:"check" yaml:"check"`
	Success int `json:"success" yaml:"success"`
	Failure int `json:"failure" yaml:"failure"`
}

// Host - host structure
type Host struct {
	ID   string   `json:"id" yaml:"id"`
	Type HostType `json:"type" yaml:"-"`

	Name string   `json:"name" yaml:"name"`
	Tags []string `json:"-" yaml:"tags,omitempty"`

	Method string `json:"method" yaml:"method"`
	URL    string `json:"url" yaml:"url"`

	Retry            string `json:"retry" yaml:"retry,omitempty"`
	Timeout          string `json:"timeout" yaml:"timeout,omitempty"`
	InitialDelay     string `json:"initialDelay" yaml:"initialDelay,omitempty"`
	SuccessThreshold int    `json:"successThreshold" yaml:"successThreshold,omitempty"`
	FailureThreshold int    `json:"failureThreshold" yaml:"failureThreshold,omitempty"`

	Success *Success          `json:"success,omitempty" yaml:"success,omitempty"`
	History *HistoryCounts    `json:"history,omitempty" yaml:"history,omitempty"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`

	RetryInterval        time.Duration `json:"-" yaml:"-"`
	TimeoutInterval      time.Duration `json:"-" yaml:"-"`
	InitialDelayInterval time.Duration `json:"-" yaml:"-"`
}

// Status - checking if provided code present in the success code list and body is equal
func (h *Host) Status(code int, b []byte) bool {
	ok := false
	for _, v := range h.Success.Code {
		if v == code {
			ok = true
		}
	}

	if ok && h.Success.Body != nil {
		ok = bytes.Compare([]byte(*h.Success.Body), b) == 0
	}

	return ok
}

// String - returns a name if available, otherwise returns the url
func (h *Host) String() string {
	if h.Name == "" {
		return h.URL
	}
	return h.Name
}

// Hash - returns some unique string per host (name + url)
func (h *Host) Hash() string {
	return fmt.Sprintf("%s_%s", h.Name, h.URL)
}

// GetType - return a host type based on url
func (h *Host) GetType() HostType {
	if strings.HasPrefix(h.URL, "mongodb://") {
		return MongoType
	}
	return HttpType
}
