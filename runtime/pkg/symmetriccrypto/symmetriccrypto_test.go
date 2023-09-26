package symmetriccrypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	enc, err := NewEphemeralEncoder(32)
	require.NoError(t, err)

	data := []byte("Hello, World!")

	cipher, err := enc.Encrypt(data)
	require.NoError(t, err)
	plain, err := enc.Decrypt(cipher)
	require.NoError(t, err)

	require.Equal(t, data, plain)
}
