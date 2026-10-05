package upgrade

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionUpToDate(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
		wantErr bool
	}{
		{name: "same version", current: "v0.90.2", latest: "v0.90.2", want: true},
		{name: "same version without v prefix", current: "0.90.2", latest: "v0.90.2", want: true},
		{name: "older patch", current: "v0.90.1", latest: "v0.90.2", want: false},
		{name: "older minor", current: "v0.89.5", latest: "v0.90.0", want: false},
		{name: "newer than latest", current: "v0.91.0", latest: "v0.90.2", want: true},
		{name: "prerelease is older than release", current: "v0.90.2-rc1", latest: "v0.90.2", want: false},
		{name: "invalid current", current: "dev", latest: "v0.90.2", wantErr: true},
		{name: "invalid latest", current: "v0.90.2", latest: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := versionUpToDate(tt.current, tt.latest)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
