package blob

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/memblob"
)

func TestDownload(t *testing.T) {
	bucket := newTestBucket(t, map[string]string{
		"2020/01/01/aata.txt": "hello",
		"2020/01/02/bata.txt": "world",
		"2020/02/03/cata.txt": "writing",
		"2020/02/04/data.txt": "test",
	})

	tests := []struct {
		name        string
		glob        string
		wantContent []string
		wantErr     bool
	}{
		{
			name:        "single file found",
			glob:        "2020/01/01/aata.txt",
			wantContent: []string{"hello"},
			wantErr:     false,
		},
		{
			name:        "recursive glob",
			glob:        "2020/**/*.txt",
			wantContent: []string{"hello", "world", "writing", "test"},
			wantErr:     false,
		},
		{
			name:        "non recursive glob",
			glob:        "2020/0?/0[1-3]/{a,b}ata.txt",
			wantContent: []string{"hello", "world"},
			wantErr:     false,
		},
		{
			name:        "glob absent",
			glob:        "2020/**/*.csv",
			wantContent: []string{},
			wantErr:     false,
		},
		{
			name:        "single file",
			glob:        "2020/01/01/aata.txt",
			wantContent: []string{"hello"},
			wantErr:     false,
		},
		{
			name:        "single file not found",
			glob:        "2020/01/01/aata.csv",
			wantContent: []string{},
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it, err := bucket.Download(context.Background(), &DownloadOptions{
				Glob:    tt.glob,
				TempDir: t.TempDir(),
			})
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			defer it.Close()

			var contents []string
			for {
				next, err := it.Next(context.Background())
				if errors.Is(err, io.EOF) {
					break
				}
				if tt.wantErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				for _, path := range next {
					data, err := os.ReadFile(path)
					require.NoError(t, err)
					contents = append(contents, string(data))
				}
			}
		})
	}
}

func newTestBucket(t *testing.T, files map[string]string) *Bucket {
	ctx := context.Background()
	underlying, err := blob.OpenBucket(ctx, "mem://")
	require.NoError(t, err)
	for key, value := range files {
		require.NoError(t, underlying.WriteAll(ctx, key, []byte(value), nil))
	}
	bucket, err := NewBucket(underlying, zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, bucket.Close())
	})
	return bucket
}
