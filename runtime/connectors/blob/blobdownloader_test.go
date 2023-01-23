package blob

import (
	"context"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/memblob"
)

var filesData = map[string][]byte{
	"2020/01/01/aata.txt": []byte("hello"),
	"2020/01/02/bata.txt": []byte("world"),
	"2020/02/03/cata.txt": []byte("writing"),
	"2020/02/04/data.txt": []byte("test"),
}

func TestFetchFileNames(t *testing.T) {
	bucket, err := prepareBucket()
	require.NoError(t, err)

	extractConfigs, err := NewExtractConfigs(nil)
	require.NoError(t, err)
	type args struct {
		ctx         context.Context
		bucket      *blob.Bucket
		config      FetchConfigs
		globPattern string
		bucketPath  string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]struct{}
		wantErr bool
	}{
		{
			name:    "single file found",
			args:    args{context.Background(), bucket, FetchConfigs{Extract: extractConfigs}, "2020/01/01/aata.txt", "mem://"},
			want:    map[string]struct{}{"hello": {}},
			wantErr: false,
		},
		{
			name:    "single file absent",
			args:    args{context.Background(), bucket, FetchConfigs{Extract: extractConfigs}, "2020/01/01/eata.txt", "mem://"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "recursive glob",
			args:    args{context.Background(), bucket, FetchConfigs{Extract: extractConfigs}, "2020/**/*.txt", "mem://"},
			want:    map[string]struct{}{"hello": {}, "world": {}, "writing": {}, "test": {}},
			wantErr: false,
		},
		{
			name:    "non recursive glob",
			args:    args{context.Background(), bucket, FetchConfigs{Extract: extractConfigs}, "2020/0?/0[1-3]/{a,b}ata.txt", "mem://"},
			want:    map[string]struct{}{"hello": {}, "world": {}},
			wantErr: false,
		},
		{
			name:    "glob absent",
			args:    args{context.Background(), bucket, FetchConfigs{Extract: extractConfigs}, "2020/**/*.csv", "mem://"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "total size limit",
			args:    args{context.Background(), bucket, FetchConfigs{GlobMaxTotalSize: 1, Extract: extractConfigs}, "2020/**", "mem://"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "max match limit",
			args:    args{context.Background(), bucket, FetchConfigs{GlobMaxObjectsMatched: 1, Extract: extractConfigs}, "2020/**", "mem://"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "max list limit",
			args:    args{context.Background(), bucket, FetchConfigs{GlobMaxObjectsListed: 1, Extract: extractConfigs}, "2020/**", "mem://"},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchFileNames(tt.args.ctx, tt.args.bucket, tt.args.config, tt.args.globPattern, tt.args.bucketPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchFileNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			require.Equal(t, len(tt.want), len(got))
			for _, path := range got {
				data, _ := os.ReadFile(path)
				strContent := string(data)
				if _, ok := tt.want[strContent]; !ok {
					t.Errorf("file with data %v not part of glob", strContent)
					return
				}
			}
		})
	}
}

func TestFetchFileNamesWithParitionLimits(t *testing.T) {
	bucket, err := prepareBucket()
	require.NoError(t, err)

	type args struct {
		ctx         context.Context
		bucket      *blob.Bucket
		config      FetchConfigs
		globPattern string
		bucketPath  string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]struct{}
		wantErr bool
	}{
		{
			name: "listing head limits",
			args: args{context.Background(),
				bucket,
				FetchConfigs{Extract: &ExtractConfigs{Partition: ExtractOptions{Strategy: "head", Size: 2}, Row: ExtractOptions{Strategy: NONE, Size: math.MaxInt64}}},
				"2020/**",
				"mem://",
			},
			want:    map[string]struct{}{"hello": {}, "world": {}},
			wantErr: false,
		},
		{
			name: "listing tail limits",
			args: args{context.Background(),
				bucket,
				FetchConfigs{Extract: &ExtractConfigs{Partition: ExtractOptions{Strategy: "tail", Size: 2}, Row: ExtractOptions{Strategy: NONE, Size: math.MaxInt64}}},
				"2020/**",
				"mem://",
			},
			want:    map[string]struct{}{"test": {}, "writing": {}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchFileNames(tt.args.ctx, tt.args.bucket, tt.args.config, tt.args.globPattern, tt.args.bucketPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchFileNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			require.Equal(t, len(tt.want), len(got))
			for _, path := range got {
				data, _ := os.ReadFile(path)
				strContent := string(data)
				if _, ok := tt.want[strContent]; !ok {
					t.Errorf("file with data %v not part of glob", strContent)
					return
				}
			}
		})
	}
}

func prepareBucket() (*blob.Bucket, error) {
	ctx := context.Background()
	bucket, err := blob.OpenBucket(ctx, "mem://")
	if err != nil {
		return nil, err
	}
	for key, value := range filesData {
		if err := bucket.WriteAll(ctx, key, value, nil); err != nil {
			return nil, err
		}
	}
	return bucket, nil
}
