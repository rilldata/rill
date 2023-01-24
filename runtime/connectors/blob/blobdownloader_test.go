package blob

import (
	"context"
	"math"
	"os"
	"reflect"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
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
			it, err := NewIterator(tt.args.ctx, tt.args.bucket, tt.args.config, tt.args.globPattern, tt.args.bucketPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchFileNames() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			paths := make([]string, 0)
			defer fileutil.ForceRemoveFiles(paths)
			for it.HasNext() {
				next, err := it.NextBatch(tt.args.ctx, 1)
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
				FetchConfigs{Extract: &ExtractPolicy{Partition: ExtractConfig{Strategy: "head", Size: 2}, Row: ExtractConfig{Strategy: NONE, Size: math.MaxInt64}}},
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
				FetchConfigs{Extract: &ExtractPolicy{Partition: ExtractConfig{Strategy: "tail", Size: 2}, Row: ExtractConfig{Strategy: NONE, Size: math.MaxInt64}}},
				"2020/**",
				"mem://",
			},
			want:    map[string]struct{}{"test": {}, "writing": {}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it, err := NewIterator(tt.args.ctx, tt.args.bucket, tt.args.config, tt.args.globPattern, tt.args.bucketPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchFileNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			paths := make([]string, 0)
			defer fileutil.ForceRemoveFiles(paths)
			for it.HasNext() {
				next, err := it.NextBatch(tt.args.ctx, 1)
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

func TestNewExtractConfigs(t *testing.T) {
	tests := []struct {
		name    string
		input   *runtimev1.Source_ExtractPolicy
		want    *ExtractPolicy
		wantErr bool
	}{
		{
			name:    "nil input",
			input:   nil,
			want:    &ExtractPolicy{Partition: ExtractConfig{Strategy: NONE}, Row: ExtractConfig{Strategy: NONE}},
			wantErr: false,
		},
		{
			name:    "parse row",
			input:   &runtimev1.Source_ExtractPolicy{Row: &runtimev1.Source_ExtractPolicy_ExtractConfig{Strategy: "tail", Size: "23 KB"}},
			want:    &ExtractPolicy{Partition: ExtractConfig{Strategy: NONE}, Row: ExtractConfig{Strategy: TAIL, Size: 23552}},
			wantErr: false,
		},
		{
			name:    "parse partition",
			input:   &runtimev1.Source_ExtractPolicy{Partition: &runtimev1.Source_ExtractPolicy_ExtractConfig{Strategy: "head", Size: "23"}},
			want:    &ExtractPolicy{Partition: ExtractConfig{Strategy: HEAD, Size: 23}, Row: ExtractConfig{Strategy: NONE}},
			wantErr: false,
		},
		{
			name:    "parse both",
			input:   &runtimev1.Source_ExtractPolicy{Partition: &runtimev1.Source_ExtractPolicy_ExtractConfig{Strategy: "tail", Size: "23"}, Row: &runtimev1.Source_ExtractPolicy_ExtractConfig{Strategy: "tail", Size: "512 B"}},
			want:    &ExtractPolicy{Partition: ExtractConfig{Strategy: TAIL, Size: 23}, Row: ExtractConfig{Strategy: TAIL, Size: 512}},
			wantErr: false,
		},
		{
			name:    "more examples",
			input:   &runtimev1.Source_ExtractPolicy{Partition: &runtimev1.Source_ExtractPolicy_ExtractConfig{Strategy: "tail", Size: "23"}, Row: &runtimev1.Source_ExtractPolicy_ExtractConfig{Strategy: "tail", Size: "23 gb"}},
			want:    &ExtractPolicy{Partition: ExtractConfig{Strategy: TAIL, Size: 23}, Row: ExtractConfig{Strategy: TAIL, Size: 23 * 1024 * 1024 * 1024}},
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   &runtimev1.Source_ExtractPolicy{Partition: &runtimev1.Source_ExtractPolicy_ExtractConfig{Strategy: "tail", Size: "23"}, Row: &runtimev1.Source_ExtractPolicy_ExtractConfig{Strategy: "tail", Size: "23%"}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewExtractConfigs(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewExtractConfigs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewExtractConfigs() = %v, want %v", got, tt.want)
			}
		})
	}
}
