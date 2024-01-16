package salesforce

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEndpoint(t *testing.T) {
	e, err := endpoint(authenticationOptions{Endpoint: "login.salesforce.com"})
	require.NoError(t, err)
	require.Equal(t, "https://login.salesforce.com", e)

	e, err = endpoint(authenticationOptions{Endpoint: "example.my.salesforce.com"})
	require.NoError(t, err)
	require.Equal(t, "https://example.my.salesforce.com", e)

	e, err = endpoint(authenticationOptions{Endpoint: "https://login.salesforce.com"})
	require.NoError(t, err)
	require.Equal(t, "https://login.salesforce.com", e)

	e, err = endpoint(authenticationOptions{Endpoint: "https://example.my.salesforce.com"})
	require.NoError(t, err)
	require.Equal(t, "https://example.my.salesforce.com", e)
}
