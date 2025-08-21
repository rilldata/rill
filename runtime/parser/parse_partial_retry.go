package parser

import (
	"fmt"
	"regexp"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

const (
	maxRetryAttempts = 10
	maxRetryDelay    = 300 // 5 minutes
)

// RetryYAML defines the YAML structure for retry configuration.
type RetryYAML struct {
	// Retry attempts
	Attempts *uint32 `yaml:"attempts" mapstructure:"attempts"`
	// Delay between retries
	Delay *uint32 `yaml:"delay" mapstructure:"delay"`
	// Enable exponential backoff
	ExponentialBackoff *bool `yaml:"exponential_backoff" mapstructure:"exponential_backoff"`
	// Errors to match
	IfErrorMatches []string `yaml:"if_error_matches" mapstructure:"if_error_matches"`
}

// parseRetryYAML converts a RetryYAML configuration into a runtime retry policy.
func (p *Parser) parseRetryYAML(raw *RetryYAML) (*runtimev1.Retry, error) {
	// No retry behavior unless explicitly configured
	if raw == nil {
		return nil, nil
	}

	// Default values when retry is configured but fields are missing
	r := &runtimev1.Retry{
		Attempts:           3,    // Default 3 attempts
		Delay:              5,    // Default 5 second delay
		ExponentialBackoff: true, // Default enable exponential backoff
		IfErrorMatches: []string{
			".*OvercommitTracker.*", // ClickHouse memory pressure
			".*Bad Gateway.*",       // HTTP 502 errors
		},
	}

	// Set attempts if provided, otherwise keep default
	if raw.Attempts != nil {
		if *raw.Attempts > maxRetryAttempts {
			return nil, fmt.Errorf("retry attempts cannot exceed the maximum of %d", maxRetryAttempts)
		}
		r.Attempts = *raw.Attempts // This allows 0
	}

	// Set delay if provided, otherwise keep default
	if raw.Delay != nil {
		if *raw.Delay > maxRetryDelay {
			return nil, fmt.Errorf("retry delay cannot exceed the maximum of %d seconds", maxRetryDelay)
		}
		r.Delay = *raw.Delay
	}

	// Always set ExponentialBackoff from raw (allows explicit false)
	if raw.ExponentialBackoff != nil {
		r.ExponentialBackoff = *raw.ExponentialBackoff
	}

	// Set error matches if provided, otherwise keep defaults
	if len(raw.IfErrorMatches) > 0 {
		// Validate regex patterns
		for _, pattern := range raw.IfErrorMatches {
			if _, err := regexp.Compile(pattern); err != nil {
				return nil, fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
			}
		}
		r.IfErrorMatches = raw.IfErrorMatches
	}

	return r, nil
}
