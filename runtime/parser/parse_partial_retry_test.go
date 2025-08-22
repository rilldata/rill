package parser

import (
	"reflect"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

func TestParseRetryYAML(t *testing.T) {
	tests := []struct {
		name    string
		input   *RetryYAML
		want    *runtimev1.Retry
		wantErr bool
	}{
		{
			name:    "nil input returns nil",
			input:   nil,
			want:    nil,
			wantErr: false,
		},
		{
			name: "valid custom input",
			input: &RetryYAML{
				Attempts:           uint32Ptr(5),
				Delay:              stringPtr("10s"),
				ExponentialBackoff: boolPtr(false),
				IfErrorMatches:     []string{".*Timeout.*"},
			},
			want: &runtimev1.Retry{
				Attempts:           5,
				Delay:              10,
				ExponentialBackoff: false,
				IfErrorMatches:     []string{".*Timeout.*"},
			},
			wantErr: false,
		},
		{
			name: "zero attempts is valid (disables retries)",
			input: &RetryYAML{
				Attempts: uint32Ptr(0),
			},
			want: &runtimev1.Retry{
				Attempts:           0,
				Delay:              5,    // default
				ExponentialBackoff: true, // default
				IfErrorMatches: []string{
					".*OvercommitTracker.*", // Memory pressure
					".*Bad Gateway.*",       // 502 Bad Gateway
				},
			},
			wantErr: false,
		},
		{
			name: "exceeding max attempts returns error",
			input: &RetryYAML{
				Attempts: uint32Ptr(11), // Max is 10, so 11 should fail
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "zero delay is valid (immediate retry)",
			input: &RetryYAML{
				Delay: stringPtr("0s"),
			},
			want: &runtimev1.Retry{
				Attempts:           3, // default
				Delay:              0,
				ExponentialBackoff: true, // default
				IfErrorMatches: []string{
					".*OvercommitTracker.*", // Memory pressure
					".*Bad Gateway.*",       // 502 Bad Gateway
				},
			},
			wantErr: false,
		},
		{
			name: "invalid delay format returns error",
			input: &RetryYAML{
				Delay: stringPtr("invalid"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:  "empty input uses all defaults",
			input: &RetryYAML{},
			want: &runtimev1.Retry{
				Attempts:           3,    // default
				Delay:              5,    // default
				ExponentialBackoff: true, // default
				IfErrorMatches: []string{
					".*OvercommitTracker.*", // Memory pressure
					".*Bad Gateway.*",       // 502 Bad Gateway
				},
			},
			wantErr: false,
		},
		{
			name: "invalid regex pattern returns error",
			input: &RetryYAML{
				IfErrorMatches: []string{"[invalid regex"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "partial retry on specific errors",
			input: &RetryYAML{
				Attempts: uint32Ptr(3),
				Delay:    stringPtr("5s"),
				IfErrorMatches: []string{
					".*OvercommitTracker.*", // Memory pressure (fixed typo)
					".*Bad Gateway.*",       // 502 Bad Gateway
				},
			},
			want: &runtimev1.Retry{
				Attempts:           3,
				Delay:              5,
				ExponentialBackoff: true,
				IfErrorMatches: []string{
					".*OvercommitTracker.*", // Memory pressure
					".*Bad Gateway.*",       // 502 Bad Gateway
				},
			},
			wantErr: false,
		},
	}

	p := &Parser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.parseRetryYAML(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRetryYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRetryYAML() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to create a pointer to uint32
func uint32Ptr(v uint32) *uint32 {
	return &v
}

// Helper function to create a pointer to bool
func boolPtr(v bool) *bool {
	return &v
}

// Helper function to create a pointer to string
func stringPtr(v string) *string {
	return &v
}
