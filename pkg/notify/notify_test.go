package notify

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/exelban/JAM/types"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("no providers", func(t *testing.T) {
		n, err := New(context.Background(), &types.Cfg{})
		require.NoError(t, err)
		require.Empty(t, n.clients)
	})
	t.Run("init slack", func(t *testing.T) {
		n, err := New(context.Background(), &types.Cfg{
			Alerts: types.Alerts{
				Slack: &types.Slack{
					Channel: "test",
					Token:   "test",
				},
			},
		})
		require.Error(t, err)
		require.Nil(t, n)
	})
}

func TestNotify_Set(t *testing.T) {
	m := &notifyMock{
		stringFunc: func() string {
			return "mock"
		},
		sendFunc: func(str string) error {
			if strings.Contains(str, "test_ok") {
				return nil
			}
			return errors.New("error")
		},
	}

	n := &Notify{
		clients: []notify{m},
	}

	require.NoError(t, n.Set(nil, types.UP, "test_ok"))
	require.Error(t, n.Set(nil, types.UP, "error"))
}
