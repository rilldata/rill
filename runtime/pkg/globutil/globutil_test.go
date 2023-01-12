package globutil

import "testing"

func TestParseURL(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    string
		want1   string
		want2   string
		wantErr bool
	}{
		{
			name:    "valid s3",
			args:    "s3://bucket.rill-developer.io/**/path_?/{0,1,2}000016[1-2]/*.parquet",
			want:    "s3",
			want1:   "bucket.rill-developer.io",
			want2:   "**/path_?/{0,1,2}000016[1-2]/*.parquet",
			wantErr: false,
		},
		{
			name:    "invalid s3",
			args:    "s3:/bucket.rill-developer.io/**/path_?/{0,1,2}000016[1-2]/*.parquet",
			want:    "",
			want1:   "",
			want2:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := ParseURL(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseURL() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseURL() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("ParseURL() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
