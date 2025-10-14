package types

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
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
	Type HostType `json:"type" yaml:"type"`

	Name        *string `json:"name,omitempty" yaml:"name,omitempty"`
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	Group       *string `json:"group,omitempty" yaml:"group,omitempty"`

	Method string `json:"method,omitempty" yaml:"method,omitempty"`
	URL    string `json:"url" yaml:"url"`

	Interval        *time.Duration `json:"interval" yaml:"interval,omitempty"` // minimum 1s
	TimeoutInterval *time.Duration `json:"timeout" yaml:"timeout,omitempty"`
	InitialDelay    *time.Duration `json:"initialDelay" yaml:"initialDelay,omitempty"`

	SuccessThreshold int `json:"successThreshold,omitempty" yaml:"successThreshold,omitempty"`
	FailureThreshold int `json:"failureThreshold,omitempty" yaml:"failureThreshold,omitempty"`

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
	hash := hasher.Sum(nil)
	return base64.URLEncoding.EncodeToString(hash)[:6]
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
	if h.Type != "" {
		return h.Type
	}

	if strings.HasPrefix(h.URL, "mongodb://") {
		return MongoType
	}
	if !strings.Contains(h.URL, "http://") && !strings.Contains(h.URL, "https://") && isIPv4(h.URL) {
		return ICMPType
	}

	return HttpType
}

// SecureURL - returns a secure url that can be used in logs or alerts. It will hide the password if present.
func (h *Host) SecureURL() string {
	url := h.URL
	if strings.HasPrefix(url, "mongodb://") {
		if strings.Contains(url, "@") {
			parts := strings.Split(url, "@")
			creds := strings.Split(parts[0], ":")
			if len(creds) == 3 {
				creds[2] = "*****"
				url = strings.Join(creds, ":") + "@" + parts[1]
			}
		}
		return url
	}
	return url
}

func isIPv4(host string) bool {
	parts := strings.Split(host, ".")

	if len(parts) != 4 {
		return false
	}

	for _, x := range parts {
		i, err := strconv.Atoi(x)
		if err != nil {
			return false
		}
		if i < 0 || i > 255 {
			return false
		}
	}
	return true
}
