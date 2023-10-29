package engine

import (
	"github.com/exelban/uptime/types"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

var r = types.HttpResponse{}

func TestLocal_Checks(t *testing.T) {
	h := NewLocal(&types.HistoryCounts{})
	require.Empty(t, h.Checks())

	num := rand.Intn(128-24) + 24
	for i := 0; i < num; i++ {
		h.checks = append(h.checks, r)
	}

	require.Len(t, h.Checks(), num)
}

func TestLocal_Add(t *testing.T) {
	t.Run("empty history count", func(t *testing.T) {
		h := NewLocal(&types.HistoryCounts{})
		require.Empty(t, h.Checks())
		h.Add(r)
		require.Empty(t, h.Checks())
	})

	t.Run("random (ok)", func(t *testing.T) {
		num := rand.Intn(128-24) + 24

		h := NewLocal(&types.HistoryCounts{
			Check: num,
		})

		for i := 0; i < num; i++ {
			h.Add(r)
		}

		require.Equal(t, num, len(h.Checks()))
	})
	t.Run("random (less)", func(t *testing.T) {
		num := rand.Intn(128-24) + 24
		limit := num - 10

		h := NewLocal(&types.HistoryCounts{
			Check: limit,
		})

		for i := 0; i < num; i++ {
			h.Add(r)
		}

		require.Equal(t, limit, len(h.Checks()))
	})
}

func TestLocal_SetStatus(t *testing.T) {
	t.Run("empty history count", func(t *testing.T) {
		h := NewLocal(&types.HistoryCounts{
			Check: 10,
		})
		h.Add(r)

		require.Empty(t, h.success)
		h.SetStatus(types.UP)
		require.Empty(t, h.success)

		require.Empty(t, h.failure)
		h.SetStatus(types.DOWN)
		require.Empty(t, h.failure)
	})

	t.Run("random (ok)", func(t *testing.T) {
		upNum := rand.Intn(128-24) + 24
		downNum := rand.Intn(128-24) + 24

		h := NewLocal(&types.HistoryCounts{
			Check:   1,
			Success: upNum,
			Failure: downNum,
		})
		h.Add(r)

		for i := 0; i < upNum; i++ {
			h.SetStatus(types.UP)
		}
		require.Equal(t, upNum, len(h.success))

		for i := 0; i < downNum; i++ {
			h.SetStatus(types.DOWN)
		}
		require.Equal(t, downNum, len(h.failure))
	})
	t.Run("random (less)", func(t *testing.T) {
		upNum := rand.Intn(128-24) + 24
		downNum := rand.Intn(128-24) + 24
		upLimit := upNum - 10
		downLimit := downNum - 10

		h := NewLocal(&types.HistoryCounts{
			Check:   1,
			Success: upLimit,
			Failure: downLimit,
		})
		h.Add(r)

		for i := 0; i < upNum; i++ {
			h.SetStatus(types.UP)
		}
		require.Equal(t, upLimit, len(h.success))

		for i := 0; i < downNum; i++ {
			h.SetStatus(types.DOWN)
		}
		require.Equal(t, downLimit, len(h.failure))
	})
}
