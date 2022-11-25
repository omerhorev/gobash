package gobash

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNextUnescaped(t *testing.T) {
	require.Equal(t, 0, nextUnescaped("abc\\d\\", 'a'))
	require.Equal(t, 1, nextUnescaped("abc\\d\\", 'b'))
	require.Equal(t, 2, nextUnescaped("abc\\d\\", 'c'))
	require.Equal(t, -1, nextUnescaped("abc\\d\\", 'd'))
	require.Equal(t, -1, nextUnescaped("abc\\d\\", 'e'))

	require.Equal(t, 0, nextUnescaped("abc\\d\\", 'a'))
	require.Equal(t, 1, nextUnescaped("abc\\d\\", 'b'))
	require.Equal(t, 2, nextUnescaped("abc\\d\\", 'c'))
	require.Equal(t, -1, nextUnescaped("abc\\d\\", 'd'))
	require.Equal(t, -1, nextUnescaped("abc\\d\\", 'e'))
}
