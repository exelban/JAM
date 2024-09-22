package store

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/exelban/JAM/types"
	bolt "go.etcd.io/bbolt"
	"sort"
	"time"
)

type Bolt struct {
	conn *bolt.DB
}

func NewBolt(ctx context.Context, path string) (*Bolt, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open bolt database: %w", err)
	}

	return &Bolt{
		conn: db,
	}, nil
}

// Close closes the store.
func (b *Bolt) Close() error {
	return b.conn.Close()
}

// Add adds a new response to the history of the given ID.
func (b *Bolt) Add(ctx context.Context, id string, r *types.HttpResponse) error {
	err := b.conn.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(id))
		if err != nil {
			return err
		}
		data, err := json.Marshal(r)
		if err != nil {
			return err
		}
		key := make([]byte, 8)
		binary.LittleEndian.PutUint64(key, uint64(r.Timestamp.UTC().Unix()))
		return bucket.Put(key, data)
	})

	return err
}

// Keys returns the keys of the store.
func (b *Bolt) Keys(ctx context.Context) ([]string, error) {
	keys := []string{}

	return keys, b.conn.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			keys = append(keys, string(name))
			return nil
		})
	})
}

// History returns the history of the given ID.
func (b *Bolt) History(ctx context.Context, id string, limit int) ([]*types.HttpResponse, error) {
	res := []*types.HttpResponse{}

	err := b.conn.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(id))
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(k, v []byte) error {
			var r types.HttpResponse
			if err := json.Unmarshal(v, &r); err != nil {
				return err
			}
			res = append(res, &r)
			return nil
		})
	})
	sort.Slice(res, func(i, j int) bool {
		return res[i].Timestamp.Before(res[j].Timestamp)
	})

	if limit > 0 && len(res) > limit {
		res = res[len(res)-limit:]
	}

	return res, err
}

// Delete deletes the given keys from the history of the given ID.
func (b *Bolt) Delete(ctx context.Context, id string, keys []time.Time) error {
	return b.conn.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(id))
		if bucket == nil {
			return nil
		}

		for _, key := range keys {
			k := make([]byte, 8)
			binary.LittleEndian.PutUint64(k, uint64(key.UTC().Unix()))
			if err := bucket.Delete(k); err != nil {
				return err
			}
		}

		return nil
	})
}
