package types

import "time"

type Service struct {
	Name      string
	Status    StatusType
	LastCheck string
	Checks    map[string]bool
	Success   []time.Time
	Failure   []time.Time
	Tags      []struct {
		Name  string
		Color string
	}
}
