package types

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Success struct {
	Code []int   `json:"code" yaml:"code"`
	Body *string `json:"body" yaml:"body"`
}

// Host - host structure
type Host struct {
	ID   string   `json:"id" yaml:"-"`
	Type HostType `json:"type" yaml:"-"`

	Name        *string  `json:"name,omitempty" yaml:"name,omitempty"`
	Description *string  `json:"description,omitempty" yaml:"description,omitempty"`
	Group       *string  `json:"group,omitempty" yaml:"group,omitempty"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`

	Method string `json:"method,omitempty" yaml:"method,omitempty"`
	URL    string `json:"url" yaml:"url"`

	Interval        *time.Duration `json:"interval" yaml:"interval,omitempty"` // minimum 1s
	TimeoutInterval *time.Duration `json:"timeout" yaml:"timeout,omitempty"`
	InitialDelay    *time.Duration `json:"initialDelay" yaml:"initialDelay,omitempty"`

	SuccessThreshold *int `json:"successThreshold,omitempty" yaml:"successThreshold,omitempty"`
	FailureThreshold *int `json:"failureThreshold,omitempty" yaml:"failureThreshold,omitempty"`

	Conditions *Success          `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	Headers    map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`

	Alerts []string `json:"alerts,omitempty" yaml:"alerts,omitempty"`

	Hidden bool `json:"hidden" yaml:"hidden"` // acceptable only if group is defined

	Index int `json:"-" yaml:"-"`
}

var ErrHostNotFound = errors.New("host not found")

// GenerateID - returns a host id based on the url hash
func (h *Host) GenerateID() string {
	hasher := md5.New()
	input := []byte(h.URL)
	if h.Group != nil {
		input = append(input, []byte(*h.Group)...)
	}
	hasher.Write(input)
	return hex.EncodeToString(hasher.Sum(nil))
}

// Status - checking if provided code present in the success code list and body is equal
func (h *Host) Status(code int, b []byte) bool {
	ok := false
	if h.Conditions == nil {
		h.Conditions = &Success{
			Code: []int{http.StatusOK},
		}
	}

	for _, v := range h.Conditions.Code {
		if v == code {
			ok = true
		}
	}

	if ok && h.Conditions.Body != nil {
		ok = bytes.Compare([]byte(*h.Conditions.Body), b) == 0
	}

	return ok
}

// String - returns a name if available, otherwise returns the url
func (h *Host) String() string {
	if h.Name == nil {
		return h.URL
	}
	return fmt.Sprintf("%s (%s)", *h.Name, h.URL)
}

// GetType - return a host type based on url
func (h *Host) GetType() HostType {
	if strings.HasPrefix(h.URL, "mongodb://") {
		return MongoType
	}
	return HttpType
}
