package notify

import (
	"errors"
	"github.com/exelban/cheks/types"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("no providers", func(t *testing.T) {
		n, err := New(&types.Cfg{})
		require.NoError(t, err)
		require.Empty(t, n.clients)
	})
	t.Run("init slack", func(t *testing.T) {
		n, err := New(&types.Cfg{
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

	require.NoError(t, n.Set(types.UP, "test_ok"))
	require.Error(t, n.Set(types.UP, "error"))
}
