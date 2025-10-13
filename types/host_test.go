package types

import (
	"crypto/md5"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
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

	t.Run("url only", func(t *testing.T) {
		h := Host{
			URL: url,
		}
		hasher := md5.New()
		hasher.Write([]byte(url))
		hash := hasher.Sum(nil)
		expected := base64.URLEncoding.EncodeToString(hash)[:6]
		require.Equal(t, expected, h.GenerateID())
	})

	t.Run("url and group", func(t *testing.T) {
		h := Host{
			URL:   url,
			Group: &group,
		}
		hasher := md5.New()
		input := append([]byte(url), []byte(group)...)
		hasher.Write(input)
		hash := hasher.Sum(nil)
		expected := base64.URLEncoding.EncodeToString(hash)[:6]
		require.Equal(t, expected, h.GenerateID())
	})
}
