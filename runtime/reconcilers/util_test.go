package reconcilers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExploreNameFromAnnotations(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		fallback    string
		want        string
	}{
		{
			name:        "explicit explore annotation",
			annotations: map[string]string{"explore": "my_explore"},
			fallback:    "fallback",
			want:        "my_explore",
		},
		{
			name:        "explicit explore takes precedence over web_open_path",
			annotations: map[string]string{"explore": "my_explore", "web_open_path": "/explore/other"},
			fallback:    "fallback",
			want:        "my_explore",
		},
		{
			name:        "web_open_path without encoding",
			annotations: map[string]string{"web_open_path": "/explore/my_explore"},
			fallback:    "fallback",
			want:        "my_explore",
		},
		{
			name:        "web_open_path with percent encoding",
			annotations: map[string]string{"web_open_path": "/explore/publisher%20overview%20explore"},
			fallback:    "fallback",
			want:        "publisher overview explore",
		},
		{
			name:        "web_open_path with trailing slash",
			annotations: map[string]string{"web_open_path": "/explore/my_explore/"},
			fallback:    "fallback",
			want:        "my_explore",
		},
		{
			name:        "web_open_path with encoding and trailing slash",
			annotations: map[string]string{"web_open_path": "/explore/publisher%20overview/"},
			fallback:    "fallback",
			want:        "publisher overview",
		},
		{
			name:        "non-explore web_open_path falls back",
			annotations: map[string]string{"web_open_path": "/canvas/my_canvas"},
			fallback:    "fallback",
			want:        "fallback",
		},
		{
			name:        "no annotations falls back",
			annotations: map[string]string{},
			fallback:    "fallback",
			want:        "fallback",
		},
		{
			name:        "nil annotations falls back",
			annotations: nil,
			fallback:    "fallback",
			want:        "fallback",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exploreNameFromAnnotations(tt.annotations, tt.fallback)
			require.Equal(t, tt.want, got)
		})
	}
}
