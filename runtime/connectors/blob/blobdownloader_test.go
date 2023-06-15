package blob

import (
	"context"
	"os"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/memblob"
)

var filesData = map[string][]byte{
	"2020/01/01/aata.txt": []byte("hello"),
	"2020/01/02/bata.txt": []byte("world"),
	"2020/02/03/cata.txt": []byte("writing"),
	"2020/02/04/data.txt": []byte("test"),
}

const TenGB = 10 * 1024 * 1024

func TestFetchFileNames(t *testing.T) {
	type args struct {
		ctx    context.Context
		bucket *blob.Bucket
		opt    Options
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]struct{}
		wantErr bool
	}{
		{
			name:    "single file found",
			args:    args{context.Background(), prepareBucket(t), Options{GlobPattern: "2020/01/01/aata.txt", StorageLimitInBytes: TenGB}},
			want:    map[string]struct{}{"hello": {}},
			wantErr: false,
		},
		{
			name:    "recursive glob",
			args:    args{context.Background(), prepareBucket(t), Options{GlobPattern: "2020/**/*.txt", StorageLimitInBytes: TenGB}},
			want:    map[string]struct{}{"hello": {}, "world": {}, "writing": {}, "test": {}},
			wantErr: false,
		},
		{
			name:    "non recursive glob",
			args:    args{context.Background(), prepareBucket(t), Options{GlobPattern: "2020/0?/0[1-3]/{a,b}ata.txt", StorageLimitInBytes: TenGB}},
			want:    map[string]struct{}{"hello": {}, "world": {}},
			wantErr: false,
		},
		{
			name:    "glob absent",
			args:    args{context.Background(), prepareBucket(t), Options{GlobPattern: "2020/**/*.csv", StorageLimitInBytes: TenGB}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "total size limit",
			args:    args{context.Background(), prepareBucket(t), Options{GlobMaxTotalSize: 1, GlobPattern: "2020/**", StorageLimitInBytes: TenGB}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "max match limit",
			args:    args{context.Background(), prepareBucket(t), Options{GlobMaxObjectsMatched: 1, GlobPattern: "2020/**", StorageLimitInBytes: TenGB}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "max list limit",
			args:    args{context.Background(), prepareBucket(t), Options{GlobMaxObjectsListed: 1, GlobPattern: "2020/**", StorageLimitInBytes: TenGB}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "storage limit exceeded",
			args:    args{context.Background(), prepareBucket(t), Options{GlobPattern: "2020/**", StorageLimitInBytes: 10}},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it, err := NewIterator(tt.args.ctx, tt.args.bucket, tt.args.opt, zap.NewNop())
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchFileNames() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				// verify bucket is already closed on error
				require.NotNil(t, tt.args.bucket.Close().Error())
				return
			}

			paths := make([]string, 0)
			defer fileutil.ForceRemoveFiles(paths)
			for it.HasNext() {
				next, err := it.NextBatch(8)
				require.NoError(t, err)
				paths = append(paths, next...)
			}
			require.Equal(t, len(tt.want), len(paths))
			for _, path := range paths {
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

	type args struct {
		ctx    context.Context
		bucket *blob.Bucket
		opts   Options
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
				prepareBucket(t),
				Options{ExtractPolicy: &runtimev1.Source_ExtractPolicy{FilesStrategy: runtimev1.Source_ExtractPolicy_STRATEGY_HEAD, FilesLimit: 2}, GlobPattern: "2020/**", StorageLimitInBytes: TenGB},
			},
			want:    map[string]struct{}{"hello": {}, "world": {}},
			wantErr: false,
		},
		{
			name: "listing tail limits",
			args: args{
				context.Background(),
				prepareBucket(t),
				Options{ExtractPolicy: &runtimev1.Source_ExtractPolicy{FilesStrategy: runtimev1.Source_ExtractPolicy_STRATEGY_TAIL, FilesLimit: 2}, GlobPattern: "2020/**", StorageLimitInBytes: TenGB},
			},
			want:    map[string]struct{}{"test": {}, "writing": {}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it, err := NewIterator(tt.args.ctx, tt.args.bucket, tt.args.opts, zap.NewNop())
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchFileNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				// verify bucket is already closed on error
				require.NotNil(t, tt.args.bucket.Close().Error())
				return
			}

			paths := make([]string, 0)
			defer fileutil.ForceRemoveFiles(paths)
			for it.HasNext() {
				next, err := it.NextBatch(8)
				require.NoError(t, err)
				paths = append(paths, next...)
			}

			require.Equal(t, len(tt.want), len(paths))
			for _, path := range paths {
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

func prepareBucket(t *testing.T) *blob.Bucket {
	ctx := context.Background()
	bucket, err := blob.OpenBucket(ctx, "mem://")
	require.NoError(t, err)

	for key, value := range filesData {
		require.NoError(t, bucket.WriteAll(ctx, key, value, nil))
	}
	return bucket
}
