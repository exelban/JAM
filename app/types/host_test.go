package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHost_Status(t *testing.T) {
	t.Run("code", func(t *testing.T) {
		h := Host{
			Success: &Success{
				Code: []int{1, 2, 3},
			},
		}

		require.False(t, h.Status(0, nil))
		require.True(t, h.Status(1, nil))
		require.True(t, h.Status(2, nil))
		require.True(t, h.Status(3, nil))
		require.False(t, h.Status(4, nil))
	})

	t.Run("body", func(t *testing.T) {
		str := "ok"
		h := Host{
			Success: &Success{
				Code: []int{200},
				Body: &str,
			},
		}

		require.False(t, h.Status(200, nil))
		require.False(t, h.Status(200, []byte("not ok")))
		require.True(t, h.Status(200, []byte(str)))
	})
}

func TestHost_String(t *testing.T) {
	name := Host{
		Name: "name",
	}
	url := Host{
		URL: "url",
	}

	require.Equal(t, "name", name.String())
	require.Equal(t, "url", url.String())
}
