package types

type StatusType string

const (
	Unknown StatusType = "unknown"
	UP      StatusType = "up"
	DOWN    StatusType = "down"
)
