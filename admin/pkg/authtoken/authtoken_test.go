package authtoken

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestValid(t *testing.T) {
	tkn := NewRandom(TypeUser)
	require.Equal(t, TypeUser, tkn.Type)
	require.True(t, tkn.ID != uuid.UUID{})
	require.True(t, tkn.Secret != [24]byte{})
	require.Len(t, tkn.SecretHash(), 32)

	str := tkn.String()
	require.True(t, len(str) > 60 && len(str) < 70)

	parsed, err := FromString(str)
	require.NoError(t, err)
	require.Equal(t, TypeUser, parsed.Type)
	require.Equal(t, tkn.ID, parsed.ID)
	require.Equal(t, tkn.Secret, parsed.Secret)
	require.Equal(t, tkn.SecretHash(), parsed.SecretHash())
}

func TestInvalid(t *testing.T) {
	valid := []string{
		"rill_usr_2Dws32dc2FxTThgCQjHerGM1rx9pJLCPQh5QbWjUiwpkZNkCCRrlrK",
		"rill_svc_2Dws32dc2FxTThgCQjHerGM1rx9pJLCPQh5QbWjUiwpkZNkCCRrlrK",
	}
	for _, tt := range valid {
		_, err := FromString(tt)
		require.NoError(t, err)
	}

	invalid := []string{
		"rill_foo_2Dws32dc2FxTThgCQjHerGM1rx9pJLCPQh5QbWjUiwpkZNkCCRrlrK",
		"roll_usr_2Dws32dc2FxTThgCQjHerGM1rx9pJLCPQh5QbWjUiwpkZNkCCRrlrK",
		"rill_usr_Z2Dws32dc2FxTThgCQjHerGM1rx9pJLCPQh5QbWjUiwpkZNkCCRrlrK",
		"rill_usr_",
		"rill__2Dws32dc2FxTThgCQjHerGM1rx9pJLCPQh5QbWjUiwpkZNkCCRrlrK",
		"rill_2Dws32dc2FxTThgCQjHerGM1rx9pJLCPQh5QbWjUiwpkZNkCCRrlrK",
		"_usr_2Dws32dc2FxTThgCQjHerGM1rx9pJLCPQh5QbWjUiwpkZNkCCRrlrK",
		"",
		"_",
		"__",
		"___",
	}
	for _, tt := range invalid {
		_, err := FromString(tt)
		require.Equal(t, ErrMalformed, err)
	}
}
