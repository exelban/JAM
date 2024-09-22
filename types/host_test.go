package types

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHost_Status(t *testing.T) {
	t.Run("code", func(t *testing.T) {
		h := Host{
			Conditions: &Success{
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
			Conditions: &Success{
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
	n := "name"
	name := Host{
		Name: &n,
		URL:  "url",
	}
	url := Host{
		URL: "url",
	}

	require.Equal(t, "name (url)", name.String())
	require.Equal(t, "url", url.String())
}

func TestHost_GenerateID(t *testing.T) {
	url := "url"
	group := "group"
	hasher := md5.New()

	hasher.Write([]byte(url))
	hash := hex.EncodeToString(hasher.Sum(nil))

	hasher.Reset()
	hasher.Write([]byte(fmt.Sprintf("%s%s", url, group)))
	groupHash := hex.EncodeToString(hasher.Sum(nil))

	t.Run("url only", func(t *testing.T) {
		h := Host{
			URL: url,
		}
		require.Equal(t, hash, h.GenerateID())
	})
	t.Run("url and group", func(t *testing.T) {
		h := Host{
			URL:   url,
			Group: &group,
		}
		require.Equal(t, groupHash, h.GenerateID())
	})
}
