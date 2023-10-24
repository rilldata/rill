package securetoken

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodecSimple(t *testing.T) {
	c := NewRandom()
	s := "hello world"
	tkn, err := c.Encode(s)
	require.NoError(t, err)
	var res string
	err = c.Decode(tkn, &res)
	require.NoError(t, err)
	require.Equal(t, s, res)

	err = c.Decode("invalid token", &res)
	require.Error(t, err)
}

func TestCodecComplex(t *testing.T) {
	type Complex struct {
		A string
		B int
	}

	c := NewRandom()
	v := &Complex{A: "hello world", B: 42}
	tkn, err := c.Encode(v)
	require.NoError(t, err)
	res := &Complex{}
	err = c.Decode(tkn, res)
	require.NoError(t, err)
	require.Equal(t, v, res)

	err = c.Decode("invalid token", res)
	require.Error(t, err)
}
