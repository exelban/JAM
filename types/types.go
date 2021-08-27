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
	Name   string
	Tags   []Tag
	Status Status

	Checks  []HttpResponse
	Success []HttpResponse
	Failure []HttpResponse
}

// Tag - color tag structure for Service
type Tag struct {
	Name  string
	Color string
}

// Status - status structure for Service
type Status struct {
	Value     StatusType
	Timestamp time.Time
}

// HttpResponse - response from the http dial
type HttpResponse struct {
	Timestamp time.Time
	Code      int
	Body      string

	OK     bool   `json:"-"`
	Bytes  []byte `json:"-"`
	Status bool   `json:"-"`

	DNS          time.Duration
	TLSHandshake time.Duration
	Connect      time.Duration
	TTFB         time.Duration
}
