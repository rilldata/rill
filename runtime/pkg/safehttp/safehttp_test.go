package safehttp

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSafeDialRejectsPrivateAddresses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	addr := strings.TrimPrefix(srv.URL, "http://")

	// The test server listens on 127.0.0.1, which hosted services must never reach
	_, err := NewTransport(false, time.Second).DialContext(t.Context(), "tcp", addr)
	require.ErrorContains(t, err, "private address")

	// Rill Developer may reach local servers
	conn, err := NewTransport(true, time.Second).DialContext(t.Context(), "tcp", addr)
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	require.True(t, isPrivateIP(net.ParseIP("169.254.169.254")))
	require.True(t, isPrivateIP(net.ParseIP("::1")))
	require.True(t, isPrivateIP(net.ParseIP("fd00::1")))
	require.False(t, isPrivateIP(net.ParseIP("93.184.216.34")))
}
