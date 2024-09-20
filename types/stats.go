package types

import "time"

// Point is a struct that contains the timestamp, status, and tooltip of a host. Used for the chart.
type Point struct {
	Status    StatusType
	Tooltip   *string
	Timestamp string
	TS        time.Time
}

// Chart is a struct that contains the points and intervals of a host.
type Chart struct {
	Points    []*Point
	Intervals []string
}

// Details is a struct that contains the uptime and response time of a host.
type Details struct {
	Uptime       []string
	ResponseTime []string
}

// Stat is a struct that contains the stats of a host.
type Stat struct {
	ID           string
	Name         *string
	Host         string
	Status       StatusType
	Uptime       int
	ResponseTime string
	Chart        Chart
	Hosts        []Stat
	Details      *Details

	Index int
}

// Event is a struct that contains the event ID, status, text, and timestamp.
type Event struct {
	ID        string
	Status    StatusType
	Text      string
	Timestamp string
}

// Stats is a struct that contains the stats of all hosts.
type Stats struct {
	IsHost bool
	Status StatusType
	Hosts  []Stat
	Events []Event
}
