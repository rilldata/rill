package globutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		url     *URL
		wantErr bool
	}{
		{
			name:    "valid s3",
			args:    "s3://bucket.rilldata.io/**/path_?/{0,1,2}000016[1-2]/*.parquet",
			url:     &URL{Scheme: "s3", Host: "bucket.rill.io", Path: "**/path_?/{0,1,2}000016[1-2]/*.parquet"},
			wantErr: false,
		},
		{
			name:    "invalid s3",
			args:    "s3:/bucket.rilldata.io/**/path_?/{0,1,2}000016[1-2]/*.parquet",
			url:     nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := ParseBucketURL(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			require.Equal(t, url, tt.url)
		})
	}
}
