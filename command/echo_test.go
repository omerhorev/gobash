package command

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEscape(t *testing.T) {
	require.Equal(t, `\`, doEscape(`\`))
	require.Equal(t, `\`, doEscape(`\\`))
	require.Equal(t, `\\`, doEscape(`\\\`))

	require.Equal(t, `12`, doEscape(`12\c34`))
	require.Equal(t, `12`, doEscape(`12\c`))
	require.Equal(t, ``, doEscape(`\c12`))
	require.Equal(t, ``, doEscape(`\c`))

	require.Equal(t, `!`, doEscape(`\041`))
	require.Equal(t, "\x04", doEscape(`\04`))
	require.Equal(t, "\x00", doEscape(`\0`))
	require.Equal(t, "\x00", doEscape(`\000`))
	require.Equal(t, "\x000", doEscape(`\0000`))
	require.Equal(t, "\x00"+"a", doEscape(`\0a`))

	require.Equal(t, "\x04", doEscape(`\x04`))
	require.Equal(t, "ÿ", doEscape(`\xff`))
	require.Equal(t, "ÿ a", doEscape(`\xff a`))
	require.Equal(t, "\x0fx", doEscape(`\xfx`))
}
