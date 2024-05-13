package pkce

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_generateCodeVerifier(t *testing.T) {
	for i := 0; i < 1000; i++ {
		code, err := generateCodeVerifier()
		require.NoError(t, err)
		require.NotEmpty(t, code)
		require.GreaterOrEqual(t, len(code), 43)
		require.LessOrEqual(t, len(code), 128)
		// only contains A-Z, a-z, 0-9, and the punctuation characters -._~ (hyphen, period, underscore, and tilde)
		for _, c := range code {
			require.Contains(t, charset, string(c))
		}
	}
}
