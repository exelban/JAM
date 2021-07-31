package types

import (
	"time"
)

// Host - host structure
type Host struct {
	Name string `json:"name" yaml:"name"`

	Method string `json:"method" yaml:"method"`
	URL    string `json:"url" yaml:"url"`

	Retry            string `json:"retry" yaml:"retry"`
	Timeout          string `json:"timeout" yaml:"timeout"`
	InitialDelay     string `json:"initialDelay" yaml:"initialDelay"`
	SuccessThreshold int    `json:"successThreshold" yaml:"successThreshold"`
	FailureThreshold int    `json:"failureThreshold" yaml:"failureThreshold"`
	SuccessCode      []int  `json:"successCode" yaml:"successCode"`

	RetryInterval        time.Duration `json:"-" yaml:"-"`
	TimeoutInterval      time.Duration `json:"-" yaml:"-"`
	InitialDelayInterval time.Duration `json:"-" yaml:"-"`
}

// ResponseCode - checking if provided code present in the success code list
func (h *Host) ResponseCode(code int) bool {
	ok := false
	for _, v := range h.SuccessCode {
		if v == code {
			ok = true
		}
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
