package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type slugTester struct {
	Name string `validate:"slug"`
}

func TestValidateSlug(t *testing.T) {
	require.NoError(t, Validate(&slugTester{Name: "helloworld"}))
	require.NoError(t, Validate(&slugTester{Name: "_hello-world"}))
	require.Error(t, Validate(&slugTester{Name: "hello world"}))
	require.Error(t, Validate(&slugTester{Name: "-hello"}))
	require.Error(t, Validate(&slugTester{Name: "ab"}))
}
