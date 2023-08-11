package pbutil

import (
	"crypto/md5"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWriteHash(t *testing.T) {
	m1 := map[string]interface{}{"a": 1, "b": "foo", "c": map[string]interface{}{"d": 2}, "e": time.Time{}}
	v1, err := ToValue(m1, nil)
	require.NoError(t, err)

	h1 := md5.New()
	err = WriteHash(v1, h1)
	require.NoError(t, err)
	s1 := hex.EncodeToString(h1.Sum(nil))
	require.Equal(t, "782fc01ffb6acfc3a7e807de1886bc6f", s1)

	delete(m1, "e")
	v2, err := ToValue(m1, nil)
	require.NoError(t, err)

	h2 := md5.New()
	err = WriteHash(v2, h2)
	require.NoError(t, err)
	s2 := hex.EncodeToString(h2.Sum(nil))
	require.Equal(t, "e8da53420944e1c2e993c086529c2236", s2)
}
