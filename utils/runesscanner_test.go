package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuneScanner(t *testing.T) {
	d := strings.NewReader("*#*a*cd#cd*hjk#*#")
	s := NewRunesScanner(d, []rune{'*', '#'})

	require.True(t, s.Scan())
	require.Equal(t, "a", s.Text())

	require.True(t, s.Scan())
	require.Equal(t, "cd", s.Text())

	require.True(t, s.Scan())
	require.Equal(t, "cd", s.Text())

	require.True(t, s.Scan())
	require.Equal(t, "hjk", s.Text())

	require.False(t, s.Scan())
}
