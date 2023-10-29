package store

import (
	"github.com/exelban/uptime/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("persistent", func(t *testing.T) {
		s, err := New(&types.HistoryCounts{
			Persistent: true,
		})
		require.Error(t, err)
		require.Nil(t, s)
	})

	t.Run("transient", func(t *testing.T) {
		s, err := New(&types.HistoryCounts{})
		require.NoError(t, err)
		require.NotNil(t, s)
	})
}
