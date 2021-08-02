package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_color(t *testing.T) {
	t.Run("check if color is used", func(t *testing.T) {
		selected := RandomColor()
		for _, c := range colors {
			if c.Value == selected {
				require.True(t, c.Used)
			}
		}
	})
	t.Run("default color if all used", func(t *testing.T) {
		for i := 0; i < len(colors); i++ {
			RandomColor()
		}
		require.Equal(t, "#268072", RandomColor())
		require.Equal(t, "#268072", RandomColor())
		require.Equal(t, "#268072", RandomColor())
	})
	t.Run("not allow run infinite loop", func(t *testing.T) {
		for i := 0; i < 300; i++ {
			RandomColor()
		}
	})
}
