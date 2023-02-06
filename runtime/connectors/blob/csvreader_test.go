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
		option *csvExtractOption
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
				option: &csvExtractOption{extractOption: &extractOption{strategy: runtimev1.Source_ExtractPolicy_STRATEGY_HEAD, limitInBytes: uint64(object.Size - 5)}, hasHeader: true},
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
				option: &csvExtractOption{extractOption: &extractOption{strategy: runtimev1.Source_ExtractPolicy_STRATEGY_TAIL, limitInBytes: uint64(object.Size - 5)}, hasHeader: true},
				fw:     getTempFile(t, object.Key),
			},
			want: resultTail,
		},
		{
			name: "tail strategy no header",
			args: args{
				ctx:    context.Background(),
				bucket: bucket,
				obj:    object,
				option: &csvExtractOption{extractOption: &extractOption{strategy: runtimev1.Source_ExtractPolicy_STRATEGY_TAIL, limitInBytes: uint64(object.Size - 5)}, hasHeader: false},
				fw:     getTempFile(t, object.Key),
			},
			want: testData[1:],
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
