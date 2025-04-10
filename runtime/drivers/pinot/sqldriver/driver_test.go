package sqldriver

import (
	"errors"
	"net"
	"testing"
)

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

func TestIsRetryableHTTPError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "timeout error",
			err:      &timeoutError{},
			expected: true,
		},
		{
			name:     "net timeout error",
			err:      &net.OpError{Err: &timeoutError{}},
			expected: true,
		},
		{
			name:     "HTTP 503",
			err:      errors.New("Pinot: 503 Service Unavailable"),
			expected: true,
		},
		{
			name:     "non-retryable HTTP error",
			err:      errors.New("Pinot: 400 Bad Request"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryableHTTPError(tt.err)
			if got != tt.expected {
				t.Errorf("isRetryableHTTPError() = %v, want %v", got, tt.expected)
			}
		})
	}
}
