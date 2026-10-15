package local

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafeLocalRedirect(t *testing.T) {
	s := &Server{app: &App{localURL: "http://localhost:9009", allowedOrigins: []string{"https://rill.example.com/"}}}
	fallback := "http://localhost:9009"

	require.Equal(t, fallback, s.safeLocalRedirect("", fallback))
	require.Equal(t, "/files/connectors/linear.yaml", s.safeLocalRedirect("/files/connectors/linear.yaml", fallback))
	require.Equal(t, "http://localhost:3001/ai", s.safeLocalRedirect("http://localhost:3001/ai", fallback))
	require.Equal(t, "http://127.0.0.1:9009/ai", s.safeLocalRedirect("http://127.0.0.1:9009/ai", fallback))
	require.Equal(t, "https://rill.example.com/ai", s.safeLocalRedirect("https://rill.example.com/ai", fallback))

	require.Equal(t, fallback, s.safeLocalRedirect("javascript:alert(1)", fallback))
	require.Equal(t, fallback, s.safeLocalRedirect("//evil.com/x", fallback))
	require.Equal(t, fallback, s.safeLocalRedirect("https://evil.com/", fallback))
	require.Equal(t, fallback, s.safeLocalRedirect("data:text/html,hi", fallback))
}
