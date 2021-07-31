package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHost_ResponseCode(t *testing.T) {
	h := Host{
		SuccessCode: []int{1, 2, 3},
	}

	require.False(t, h.ResponseCode(0))
	require.True(t, h.ResponseCode(1))
	require.True(t, h.ResponseCode(2))
	require.True(t, h.ResponseCode(3))
	require.False(t, h.ResponseCode(4))
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
