package config

import (
	"bytes"
	"fmt"
	"time"
)

type Success struct {
	Code []int   `json:"code" yaml:"code"`
	Body *string `json:"body" yaml:"body"`
}

type HistoryCounts struct {
	Check   int `json:"check" yaml:"check"`
	Success int `json:"success" yaml:"success"`
	Failure int `json:"failure" yaml:"failure"`
}

type StatusType string

const (
	Unknown StatusType = "unknown"
	UP      StatusType = "up"
	DOWN    StatusType = "down"
)

// Host - host structure
type Host struct {
	Name string   `json:"name" yaml:"name"`
	Tags []string `json:"tags" yaml:"tags"`

	Method string `json:"method" yaml:"method"`
	URL    string `json:"url" yaml:"url"`

	Retry            string `json:"retry" yaml:"retry"`
	Timeout          string `json:"timeout" yaml:"timeout"`
	InitialDelay     string `json:"initialDelay" yaml:"initialDelay"`
	SuccessThreshold int    `json:"successThreshold" yaml:"successThreshold"`
	FailureThreshold int    `json:"failureThreshold" yaml:"failureThreshold"`

	Success *Success          `json:"success" yaml:"success"`
	History *HistoryCounts    `json:"history" yaml:"history"`
	Headers map[string]string `json:"headers" yaml:"headers"`

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
