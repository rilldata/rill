package blob

import (
	"bytes"
	"context"
	"encoding/csv"
	"os"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/stretchr/testify/require"
	"gocloud.dev/blob"
)

var testData = [][]string{{"year", "sale"}, {"2020", "1"}, {"2021", "100"}, {"2022", "10000"}}
var resultTail = [][]string{{"year", "sale"}, {"2021", "100"}, {"2022", "10000"}}

func TestDownloadCSV(t *testing.T) {
	bucket, object, err := prepareBucketCSV()
	require.NoError(t, err)
	type args struct {
		ctx    context.Context
		bucket *blob.Bucket
		obj    *blob.ListObject
		option *extractOption
		fw     *os.File
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    [][]string
	}{
		{
			name: "head strategy",
			args: args{
				ctx:    context.Background(),
				bucket: bucket,
				obj:    object,
				option: &extractOption{strategy: runtimev1.Source_ExtractPolicy_STRATEGY_HEAD, limitInBytes: uint64(object.Size - 5)},
				fw:     getTempFile(t, object.Key),
			},
			want: testData[:len(testData)-1],
		},
		{
			name: "tail strategy",
			args: args{
				ctx:    context.Background(),
				bucket: bucket,
				obj:    object,
				option: &extractOption{strategy: runtimev1.Source_ExtractPolicy_STRATEGY_TAIL, limitInBytes: uint64(object.Size - 5)},
				fw:     getTempFile(t, object.Key),
			},
			want: resultTail,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := downloadCSV(tt.args.ctx, tt.args.bucket, tt.args.obj, tt.args.option, tt.args.fw); (err != nil) != tt.wantErr {
				t.Errorf("DownloadCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.args.fw.Close()

			r, err := os.Open(tt.args.fw.Name())
			require.NoError(t, err)

			csvReader := csv.NewReader(r)
			records, err := csvReader.ReadAll()
			require.NoError(t, err)
			require.Equal(t, tt.want, records)
		})
	}
}

func TestDownloadCSVSingleLineHead(t *testing.T) {
	ctx := context.Background()
	bucket, err := blob.OpenBucket(ctx, "mem://")
	require.NoError(t, err)
	p := []byte("1,2020")
	require.NoError(t, bucket.WriteAll(ctx, "data.csv", p, nil))

	withNewLine := []byte("1,2020\n")
	require.NoError(t, bucket.WriteAll(ctx, "withNewLine.csv", withNewLine, nil))

	for _, file := range []string{"data.csv", "withNewLine.csv"} {
		object, err := bucket.List(&blob.ListOptions{Prefix: file}).Next(ctx)
		require.NoError(t, err)

		extractOption := &extractOption{strategy: runtimev1.Source_ExtractPolicy_STRATEGY_HEAD, limitInBytes: uint64(object.Size)}
		fw := getTempFile(t, "temp.csv")
		err = downloadCSV(ctx, bucket, object, extractOption, fw)
		require.NoError(t, err)
		fw.Close()

		r, err := os.Open(fw.Name())
		require.NoError(t, err)

		csvReader := csv.NewReader(r)
		records, err := csvReader.ReadAll()
		require.NoError(t, err)
		require.Equal(t, [][]string{{"1", "2020"}}, records)
	}
}

func getTempFile(t *testing.T, name string) *os.File {
	fw, err := fileutil.OpenTempFileInDir(t.TempDir(), name)
	require.NoError(t, err)
	return fw
}

func prepareBucketCSV() (*blob.Bucket, *blob.ListObject, error) {
	ctx := context.Background()
	bucket, err := blob.OpenBucket(ctx, "mem://")
	if err != nil {
		return nil, nil, err
	}

	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	if err := w.WriteAll(testData); err != nil {
		return nil, nil, err
	}

	p := buf.Bytes()
	if err := bucket.WriteAll(ctx, "data.csv", p, nil); err != nil {
		return nil, nil, err
	}

	object, err := bucket.List(&blob.ListOptions{Prefix: "data.csv"}).Next(ctx)

	return bucket, object, err
}
