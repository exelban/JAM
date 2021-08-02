package types

type Service struct {
	Status    StatusType
	LastCheck string
	History   map[string]bool
}
