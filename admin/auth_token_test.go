package admin_test

import (
	"strings"
	"testing"

	"github.com/rilldata/rill/admin"
	"github.com/stretchr/testify/require"
)

func TestMagicAuthTokenPasswordValidation(t *testing.T) {
	for _, tc := range []struct {
		name     string
		password string
		valid    bool
	}{
		{"empty", "", false},
		{"short", "1234567", false},
		{"minimum", "12345678", true},
		{"maximum ASCII", strings.Repeat("a", 72), true},
		{"too many bytes", strings.Repeat("a", 73), false},
		{"maximum accented", strings.Repeat("é", 36), true},
		{"accented byte overflow", strings.Repeat("é", 37), false},
		{"emoji character minimum", strings.Repeat("😀", 4), false},
		{"maximum emoji", strings.Repeat("😀", 18), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := admin.ValidateMagicAuthTokenPassword(tc.password)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				_, err = admin.HashMagicAuthTokenPassword(tc.password)
				require.Error(t, err)
			}
		})
	}
}
