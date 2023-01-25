package blob

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/apache/arrow/go/v11/arrow"
	"github.com/apache/arrow/go/v11/arrow/array"
	"github.com/apache/arrow/go/v11/arrow/memory"
	"github.com/apache/arrow/go/v11/parquet"
	"github.com/apache/arrow/go/v11/parquet/compress"
	"github.com/apache/arrow/go/v11/parquet/file"
	"github.com/apache/arrow/go/v11/parquet/pqarrow"
	"github.com/c2h5oh/datasize"
	"gocloud.dev/blob"
)

var batchSize = int64(1000)

type blobObjectReader struct {
	ctx    context.Context
	bucket *blob.Bucket
	index  int64
	obj    *blob.ListObject

	// debug data
	debugMode bool
	call      int64
	bytes     int64
}

// todo :: add buffer for caching
func (f *blobObjectReader) ReadAt(p []byte, off int64) (int, error) {
	if f.debugMode {
		fmt.Printf("reading %v bytes at offset %v\n", len(p), off)
		atomic.AddInt64(&f.call, 1)
	}

	reader, err := f.bucket.NewRangeReader(f.ctx, f.obj.Key, off, int64(len(p)), nil)
	if err != nil {
		return 0, err
	}
	defer reader.Close()

	n, err := io.ReadFull(reader, p)
	if err != nil {
		return 0, err
	}
	if f.debugMode {
		atomic.AddInt64(&f.bytes, int64(n))
	}
	return n, nil
}

func (f *blobObjectReader) Read(p []byte) (int, error) {
	n, err := f.ReadAt(p, f.index)
	f.index += int64(n)
	return n, err
}

func (f *blobObjectReader) Size() int64 {
	return f.obj.Size
}

func (f *blobObjectReader) Close() error {
	if f.debugMode {
		bytes := datasize.ByteSize(f.bytes)
		fmt.Printf("made %v calls data fetched %v \n", f.call, bytes.HumanReadable())
	}
	return nil
}

func (f *blobObjectReader) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = f.index + offset
	case io.SeekEnd:
		abs = f.Size() + offset
	default:
		return 0, errors.New("bytes.Reader.Seek: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("bytes.Reader.Seek: negative position")
	}
	f.index = abs

	return abs, nil
}

func newBlobObjectReader(ctx context.Context, bucket *blob.Bucket, obj *blob.ListObject) *blobObjectReader {
	return &blobObjectReader{
		ctx:       ctx,
		bucket:    bucket,
		obj:       obj,
		debugMode: true,
	}
}

func Reverse(s []int) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func getArray(size int, rev bool) []int {
	result := make([]int, size)
	for i := 0; i < size; i++ {
		result[i] = i
	}

	if rev {
		Reverse(result)
	}
	return result
}

func estimate(reader *file.Reader, option ExtractConfig) ([]int, int64) {
	rowIndexes := getArray(reader.NumRowGroups(), option.Strategy == TAIL)

	result := make([]int, 0)
	var cumSize, rows int64
	for _, index := range rowIndexes {
		result = append(result, index)
		rowGroup := reader.RowGroup(index)
		rowGroupSize := rowGroup.ByteSize()
		rowCount := rowGroup.NumRows()

		if cumSize+rowGroupSize > option.Size {
			// taking entire rowgroup crosses allowed size
			perRowSize := rowGroupSize / rowCount
			rows += (option.Size - cumSize) / perRowSize
			return result, rows
		}
		rows += rowCount
	}
	return result, rows
}

func Download(ctx context.Context, bucket *blob.Bucket, obj *blob.ListObject, option ExtractConfig, fw *os.File) error {
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	reader := newBlobObjectReader(ctx, bucket, obj)

	props := parquet.NewReaderProperties(mem)
	props.BufferedStreamEnabled = true

	pf, err := file.NewParquetReader(reader, file.WithReadProps(props))
	if err != nil {
		return err
	}
	defer pf.Close()

	arrowReadProperties := pqarrow.ArrowReadProperties{BatchSize: batchSize, Parallel: true}
	fileReader, err := pqarrow.NewFileReader(pf, arrowReadProperties, mem)
	if err != nil {
		return err
	}

	numRowGroups := pf.NumRowGroups()
	if numRowGroups == 0 {
		return fmt.Errorf("invalid parquet")
	}

	rowIndices, rowLimit := estimate(pf, option)

	r, err := fileReader.GetRecordReader(ctx, getArray(pf.RowGroup(0).NumColumns(), false), rowIndices)
	if err != nil {
		return err
	}

	// one record has batchsize rows
	records := make([]arrow.Record, rowLimit/batchSize)
	for i := 0; i < len(records); i++ {
		// one read fetch batchsize number of rows in one call
		rec, err := r.Read()
		if err != nil {
			return err
		}

		// need to explicitly retain, else memory is reclaimed
		rec.Retain()
		records[i] = rec
	}
	defer func() {
		for _, rec := range records {
			rec.Release()
		}
	}()

	schema, err := fileReader.Schema()
	if err != nil {
		return err
	}

	table := array.NewTableFromRecords(schema, records)
	defer table.Release()

	// duck db requires root Repetitions to be required
	// keeping compressions as uncompressed since file will be immediately consumed and deleted
	wp := parquet.NewWriterProperties(
		parquet.WithRootRepetition(parquet.Repetitions.Required),
		parquet.WithCompression(compress.Codecs.Uncompressed),
		parquet.WithAllocator(mem))
	return pqarrow.WriteTable(
		table,
		fw,
		table.NumRows(),
		wp,
		pqarrow.NewArrowWriterProperties(pqarrow.WithAllocator(mem)),
	)
}
