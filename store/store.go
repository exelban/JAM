package store

import (
	"context"
	"fmt"
	"github.com/exelban/JAM/types"
	"log"
	"math/rand/v2"
	"time"
)

type Interface interface {
	Add(ctx context.Context, id string, r *types.HttpResponse) error
	Keys(ctx context.Context) ([]string, error)
	History(ctx context.Context, id string, limit int) ([]*types.HttpResponse, error)
	Delete(ctx context.Context, id string, keys []time.Time) error

	Close() error
}

func New(ctx context.Context, cfg *types.Cfg) (Interface, error) {
	var store Interface

	if cfg != nil && cfg.Storage != nil {
		path := "./jam.db"
		if cfg.Storage.Path != nil {
			path = fmt.Sprintf("%s/jam.db", *cfg.Storage.Path)
		}

		switch cfg.Storage.Type {
		case "bolt":
			s, err := NewBolt(ctx, path)
			if err != nil {
				return nil, err
			}
			store = s
		}
	} else {
		store = NewMemory(ctx)
	}

	if err := Aggregate(ctx, store); err != nil {
		return nil, fmt.Errorf("failed to aggregate: %w", err)
	}

	tk := time.NewTicker(hoursToMidnight())
	go func() {
		for {
			select {
			case <-tk.C:
				if err := Aggregate(ctx, store); err != nil {
					log.Printf("[ERROR] failed to aggregate: %v", err)
				}
				nextRun := hoursToMidnight()
				tk.Reset(nextRun)
				log.Printf("[INFO] next aggregation in %v", nextRun)
			case <-ctx.Done():
				tk.Stop()
				return
			}
		}
	}()

	return store, nil
}

func Aggregate(ctx context.Context, s Interface) error {
	log.Printf("[INFO] aggregating data")
	start := time.Now()

	keys, err := s.Keys(ctx)
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	for _, key := range keys {
		days := make(map[time.Time][]*types.HttpResponse)

		history, err := s.History(ctx, key, -1)
		if err != nil {
			return fmt.Errorf("failed to get history for %s: %w", key, err)
		}

		for _, r := range history {
			if r.Timestamp.After(today) || r.IsAggregated {
				continue
			}
			y, m, d := r.Timestamp.Date()
			day := time.Date(y, m, d, 0, 0, 0, 0, r.Timestamp.Location())
			if _, ok := days[day]; !ok {
				days[day] = make([]*types.HttpResponse, 0)
			}
			days[day] = append(days[day], r)
		}

		for day, responses := range days {
			aggregation := AggregateDay(day, responses)

			toDelete := make([]time.Time, 0)
			for _, r := range responses {
				toDelete = append(toDelete, r.Timestamp)
			}

			if err := s.Delete(ctx, key, toDelete); err != nil {
				return err
			}
			if err := s.Add(ctx, key, aggregation); err != nil {
				return err
			}
		}
	}

	log.Printf("[INFO] aggregation took %v", time.Since(start))
	return nil
}
func AggregateDay(ts time.Time, responses []*types.HttpResponse) *types.HttpResponse {
	aggregation := &types.HttpResponse{
		Timestamp:    ts,
		IsAggregated: true,
		Uptime:       0,
		Count:        len(responses),
		Time:         0,
	}
	if len(responses) == 0 {
		return aggregation
	}

	for _, r := range responses {
		if r.StatusType != types.DOWN {
			aggregation.Uptime++
		}
		aggregation.Time += r.Time
	}

	aggregation.Uptime = aggregation.Uptime / float64(len(responses))
	aggregation.Time = aggregation.Time / time.Duration(len(responses))

	if aggregation.Uptime > 0.95 {
		aggregation.StatusType = types.UP
	} else if aggregation.Uptime > 0.5 {
		aggregation.StatusType = types.DEGRADED
	} else {
		aggregation.StatusType = types.DOWN
	}

	return aggregation
}

func GenerateHistory(s Interface, start time.Time, id string) int {
	ctx := context.Background()
	now := time.Now()
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	count := rand.IntN(10000-3000) + 3000
	interval := now.Sub(start) / time.Duration(count)
	today := 0

	for i := 0; i < count; i++ {
		var status types.StatusType
		switch randInt(1, 3) {
		case 1:
			status = types.UP
		default:
			status = types.DOWN
		}
		r := &types.HttpResponse{
			Timestamp:  start.Add(interval * time.Duration(i)),
			Code:       i,
			StatusType: status,
			Time:       time.Duration(randInt(1, 100)) * time.Millisecond,
		}
		if r.Timestamp.After(day) {
			today++
		}
		if err := s.Add(ctx, id, r); err != nil {
			fmt.Printf("failed to add: %v\n", err)
		}
	}

	return today
}

func randInt(min, max int) int {
	return rand.IntN(max-min) + min
}
func hoursToMidnight() time.Duration {
	return time.Until(time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour).Add(time.Minute * 10))
}
