package config

import "time"

type Tag struct {
	Name  string
	Color string
}

type Status struct {
	Value     StatusType
	Timestamp time.Time
}

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

type Service struct {
	Name   string
	Tags   []Tag
	Status Status

	Checks  []HttpResponse
	Success []HttpResponse
	Failure []HttpResponse
}
