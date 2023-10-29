package store

import (
	"errors"
	"github.com/exelban/uptime/store/engine"
	"github.com/exelban/uptime/types"
)

type Store interface {
	Checks() []types.HttpResponse
	Success() []types.HttpResponse
	Failure() []types.HttpResponse

	Add(r types.HttpResponse)
	SetStatus(value types.StatusType)
}

func New(config *types.HistoryCounts) (Store, error) {
	if config.Persistent {
		return nil, errors.New("not implemented")
	}

	return engine.NewLocal(config), nil
}
