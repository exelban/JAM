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
	t.Run("init slack error", func(t *testing.T) {
		n, err := New(context.Background(), &types.Cfg{
			Notifications: types.Notifications{
				Slack: &types.Slack{
					Channel: "test",
					Token:   "test",
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, n)
	})
}

func TestNotify_Set(t *testing.T) {
	m := &notifyMock{
		stringFunc: func() string {
			return "mock"
		},
		sendFunc: func(subject, body string) error {
			if strings.Contains(body, "test_ok") {
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
