package types

import (
	"time"
)

// StatusType - host status type
type StatusType string

const (
	Unknown StatusType = "unknown"
	UP      StatusType = "up"
	DOWN    StatusType = "down"
)

// Service - structure for API
type Service struct {
	ID     string `json:"id"`
	Status Status `json:"status"`

	Tags []Tag `json:"tags"`

	Checks  []HttpResponse `json:"checks"`
	Success []HttpResponse `json:"success"`
	Failure []HttpResponse `json:"failure"`
}

// Tag - color tag structure for Service
type Tag struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Status - status structure for Service
type Status struct {
	Value     StatusType `json:"value"`
	Timestamp time.Time  `json:"timestamp"`
}

// HttpResponse - response from the http dial
type HttpResponse struct {
	Timestamp time.Time `json:"timestamp"`
	Time      int64     `json:"time"`
	Code      int       `json:"code"`
	Body      string    `json:"body"`

	OK     bool   `json:"-"`
	Bytes  []byte `json:"-"`
	Status bool   `json:"status"`

	DNS          time.Duration `json:"DNS"`
	TLSHandshake time.Duration `json:"TLS_handshake"`
	Connect      time.Duration `json:"connect"`
	TTFB         time.Duration `json:"TTFB"`
}

type HostType string

var (
	HttpType  HostType = "http"
	MongoType HostType = "mongo"
)
