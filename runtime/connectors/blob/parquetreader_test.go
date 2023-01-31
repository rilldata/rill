package blob

import (
	"context"
	"os"
	"testing"

	"github.com/apache/arrow/go/v11/arrow"
	"github.com/apache/arrow/go/v11/arrow/array"
	"github.com/apache/arrow/go/v11/arrow/memory"
	"github.com/apache/arrow/go/v11/parquet/file"
	"github.com/apache/arrow/go/v11/parquet/pqarrow"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/stretchr/testify/require"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
)

func TestDownload(t *testing.T) {
	bucket, object := prepareFileBucket(t)

	type args struct {
		ctx    context.Context
		bucket *blob.Bucket
		obj    *blob.ListObject
		option *extractOption
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []int32
	}{
		{
			name:    "download partial head",
			args:    args{ctx: context.Background(), bucket: bucket, obj: object, option: &extractOption{strategy: runtimev1.Source_ExtractPolicy_STRATEGY_HEAD, limitInBytes: 1000}},
			wantErr: false,
			want:    getInt32Array(1000, false),
		},
		{
			name:    "download partial tail",
			args:    args{ctx: context.Background(), bucket: bucket, obj: object, option: &extractOption{strategy: runtimev1.Source_ExtractPolicy_STRATEGY_TAIL, limitInBytes: 1000}},
			wantErr: false,
			want:    getInt32Array(2000, false)[1000:],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := fileutil.OpenTempFileInDir(t.TempDir(), "test.parquet")
			err := downloadParquet(tt.args.ctx, tt.args.bucket, tt.args.obj, tt.args.option, file)
			file.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			result := getData(t, file.Name())
			require.Equal(t, tt.want, result)
		})
	}
}

func prepareFileBucket(t *testing.T) (*blob.Bucket, *blob.ListObject) {
	ctx := context.Background()
	bucket, err := blob.OpenBucket(ctx, "mem://")
	require.NoError(t, err)

	data, err := os.ReadFile(writeParquetFile(t))
	require.NoError(t, err)

	require.NoError(t, bucket.WriteAll(ctx, "out.parquet", data, nil))

	object, err := bucket.List(&blob.ListOptions{Prefix: "out.parquet"}).Next(ctx)
	require.NoError(t, err)

	return bucket, object
}

func createArrowTable(mem memory.Allocator) arrow.Table {
	dtypeValues := make([]int32, 2000)
	for i := 0; i < 2000; i++ {
		dtypeValues[i] = int32(i)
	}

	dtype := arrow.Field{Name: "data_type", Type: arrow.PrimitiveTypes.Int32, Nullable: false}

	fieldList := []arrow.Field{dtype}

	arrsc := arrow.NewSchema(fieldList, nil)
	builders := make([]array.Builder, 0, len(fieldList))
	for _, f := range fieldList {
		bldr := array.NewBuilder(mem, f.Type)
		defer bldr.Release()
		builders = append(builders, bldr)
	}

	builders[0].(*array.Int32Builder).AppendValues(dtypeValues, make([]bool, 0))

	cols := make([]arrow.Column, 0, len(fieldList))

	for idx, field := range fieldList {
		arr := builders[idx].NewArray()
		defer arr.Release()

		chunked := arrow.NewChunked(field.Type, []arrow.Array{arr})
		defer chunked.Release()
		col := arrow.NewColumn(field, chunked)
		defer col.Release()
		cols = append(cols, *col)
	}

	return array.NewTable(arrsc, cols, int64(2000))
}

func writeParquetFile(t *testing.T) string {
	tempDir := t.TempDir()
	file, err := fileutil.OpenTempFileInDir(tempDir, "out.parquet")
	require.NoError(t, err)
	defer file.Close()

	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	tbl := createArrowTable(mem)
	defer tbl.Release()

	pqarrow.WriteTable(
		tbl,
		file,
		tbl.NumRows(),
		nil,
		pqarrow.NewArrowWriterProperties(pqarrow.WithAllocator(mem)))

	return file.Name()
}

func getData(t *testing.T, name string) []int32 {
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)

	f, err := file.OpenParquetFile(name, false)
	require.NoError(t, err)

	arrowRdr, err := pqarrow.NewFileReader(f, pqarrow.ArrowReadProperties{}, mem)
	require.NoError(t, err)

	tbl, err := arrowRdr.ReadTable(context.Background())
	require.NoError(t, err)

	defer tbl.Release()

	data := tbl.Column(0).Data()
	arr := make([]int32, 0)
	for _, chunk := range data.Chunks() {
		arr = append(arr, array.NewInt32Data(chunk.Data()).Int32Values()...)
	}
	return arr
}

func getInt32Array(size int, rev bool) []int32 {
	result := make([]int32, size)
	for i := 0; i < size; i++ {
		result[i] = int32(i)
	}
	return result
}
