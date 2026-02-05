package reconcilers

import (
	"regexp"
	"testing"
)

// TestDefaultRetryPatterns verifies that the default retry patterns in executeWithRetry
// match common transient errors from cloud storage providers and network operations.
//
// This test serves as documentation for which errors we expect to retry.
// Sources:
// - GCS: https://cloud.google.com/storage/docs/retry-strategy
// - S3: https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html
// - Go net: https://gosamples.dev/connection-reset-by-peer/
func TestDefaultRetryPatterns(t *testing.T) {
	// These patterns must match the ones in executeWithRetry (model.go)
	patterns := []string{
		// ClickHouse-specific
		".*OvercommitTracker.*",
		// HTTP errors
		".*Bad Gateway.*",           // 502
		".*Service Unavailable.*",   // 503
		".*Internal Server Error.*", // 500
		".*Gateway Timeout.*",       // 504
		"(?i).*InternalError.*",     // S3 internal error
		// Timeouts
		"(?i).*timeout.*",
		"(?i).*i/o timeout.*",
		"(?i).*TLS handshake timeout.*",
		// Connection errors
		"(?i).*connection refused.*",
		"(?i).*connection reset.*",
		"(?i).*broken pipe.*",
		"(?i).*EOF.*",
		// Network errors
		"(?i).*network.*unreachable.*",
		"(?i).*no such host.*",
		"(?i).*temporarily unavailable.*",
		// HTTP/2 stream errors
		".*stream error.*",
	}

	// Compile all patterns to verify they're valid regexes
	compiled := make([]*regexp.Regexp, len(patterns))
	for i, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			t.Fatalf("pattern %q failed to compile: %v", p, err)
		}
		compiled[i] = re
	}

	// Errors that SHOULD trigger a retry
	shouldMatch := []struct {
		name  string
		error string
	}{
		// GCS errors (https://cloud.google.com/storage/docs/retry-strategy)
		{"GCS 503 backend error", "googleapi: Error 503: Service Unavailable, backendError"},
		{"GCS 500 internal", "googleapi: Error 500: Internal Server Error"},
		{"GCS connection reset", "Get \"https://storage.googleapis.com/bucket/file\": read tcp 10.0.0.1:443: read: connection reset by peer"},

		// S3 errors (https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html)
		{"S3 SlowDown", "SlowDown: Please reduce your request rate. Service Unavailable"},
		{"S3 InternalError", "InternalError: We encountered an internal error. Please try again."},
		{"S3 503", "Service Unavailable: The server is temporarily unable to handle your request"},

		// Go net package errors (https://gosamples.dev/connection-reset-by-peer/)
		{"net connection reset", "read tcp [::1]:65244->[::1]:8080: read: connection reset by peer"},
		{"net broken pipe", "write tcp [::1]:62575->[::1]:8080: write: broken pipe"},
		{"net connection refused", "dial tcp 127.0.0.1:8080: connect: connection refused"},
		{"net i/o timeout", "dial tcp 1.2.3.4:443: i/o timeout"},
		{"net DNS failure", "dial tcp: lookup storage.googleapis.com: no such host"},

		// HTTP proxy errors
		{"nginx 502", "502 Bad Gateway"},
		{"nginx 504", "504 Gateway Timeout"},

		// TLS errors
		{"TLS timeout", "net/http: TLS handshake timeout"},

		// HTTP/2 errors
		{"HTTP/2 stream error", "stream error: stream ID 1; INTERNAL_ERROR"},

		// ClickHouse errors
		{"ClickHouse memory", "Code: 241. DB::Exception: Memory limit (total) exceeded: OvercommitTracker"},

		// Generic errors
		{"EOF during read", "unexpected EOF"},
		{"network unreachable", "dial tcp: network is unreachable"},
		{"resource temporarily unavailable", "read: resource temporarily unavailable"},
	}

	for _, tc := range shouldMatch {
		t.Run("should_match/"+tc.name, func(t *testing.T) {
			matched := false
			for _, re := range compiled {
				if re.MatchString(tc.error) {
					matched = true
					break
				}
			}
			if !matched {
				t.Errorf("error %q should match a retry pattern but didn't", tc.error)
			}
		})
	}

	// Errors that should NOT trigger a retry (permanent errors)
	shouldNotMatch := []struct {
		name  string
		error string
	}{
		{"not found", "file not found"},
		{"permission denied", "permission denied"},
		{"invalid argument", "invalid argument: column 'foo' does not exist"},
		{"S3 access denied", "AccessDenied: Access Denied"},
		{"S3 no such key", "NoSuchKey: The specified key does not exist"},
		{"S3 no such bucket", "NoSuchBucket: The specified bucket does not exist"},
		{"GCS 404", "googleapi: Error 404: Not Found"},
		{"GCS 403", "googleapi: Error 403: Access denied"},
		{"syntax error", "syntax error at or near 'SELECT'"},
	}

	for _, tc := range shouldNotMatch {
		t.Run("should_not_match/"+tc.name, func(t *testing.T) {
			for _, re := range compiled {
				if re.MatchString(tc.error) {
					t.Errorf("error %q should NOT match retry patterns but matched %q", tc.error, re.String())
					break
				}
			}
		})
	}
}
