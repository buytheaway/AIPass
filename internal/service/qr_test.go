package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashQRTokenIsStableAndDoesNotReturnRawToken(t *testing.T) {
	token := "sample-token"

	first := HashQRToken(token)
	second := HashQRToken(token)

	require.Equal(t, first, second)
	require.NotEqual(t, token, first)
	require.Len(t, first, 64)
}

func TestRandomTokenIsURLSafe(t *testing.T) {
	token, err := randomToken()

	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotContains(t, token, "+")
	require.NotContains(t, token, "/")
	require.NotContains(t, token, "=")
}
