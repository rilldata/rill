package blob

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/apache/arrow/go/v11/arrow"
	"github.com/apache/arrow/go/v11/arrow/array"
	"github.com/apache/arrow/go/v11/arrow/memory"
	"github.com/apache/arrow/go/v11/parquet"
	"github.com/apache/arrow/go/v11/parquet/compress"
	"github.com/apache/arrow/go/v11/parquet/file"
	"github.com/apache/arrow/go/v11/parquet/pqarrow"
	"gocloud.dev/blob"
)

var batchSize = int64(100)

type parquetReader struct {
	ctx    context.Context
	bucket *blob.Bucket
	index  int64
	obj    *blob.ListObject

	// buffer for caching
	buffer         []byte
	bufStartOffset int64
	bufEndOffset   int64

	// debug data
	debugMode bool
	cached    int
	call      int
}

func (f *parquetReader) WithinBuffer(start, end int64) bool {
	return f.bufStartOffset <= start && f.bufEndOffset > end
}

func (f *parquetReader) ReadInBuffer(start int64) error {
	length := int64(len(f.buffer))
	if start+length >= f.obj.Size {
		// limit end offset to object size
		length = f.obj.Size - start
	}
	reader, err := f.bucket.NewRangeReader(f.ctx, f.obj.Key, start, length, nil)
	if err != nil {
		panic(err)
	}

	defer reader.Close()

	bytes, err := reader.Read(f.buffer)
	if err != nil {
		panic(err)
	}

	f.bufStartOffset = start
	f.bufEndOffset = start + int64(bytes)
	return nil
}

// todo :: add buffer for caching
func (f *parquetReader) ReadAt(p []byte, off int64) (int, error) {
	fmt.Printf("reading %v bytes at offset %v\n", len(p), off)
	f.call++

	// end := off + int64(len(p))
	// if len(p) <= len(f.buffer) {
	// 	if !f.WithinBuffer(off, end) {
	// 		if err := f.ReadInBuffer(off); err != nil {
	// 			panic(err)
	// 		}
	// 	} else {
	// 		f.cached++
	// 	}
	// 	return copy(p, f.buffer), nil
	// }

	reader, err := f.bucket.NewRangeReader(f.ctx, f.obj.Key, off, int64(len(p)), nil)
	if err != nil {
		return 0, err
	}
	defer reader.Close()

	return reader.Read(p)
}

func (f *parquetReader) Read(p []byte) (int, error) {
	n, err := f.ReadAt(p, f.index)
	f.index += int64(n)
	return n, err
}

func (f *parquetReader) Size() int64 {
	return f.obj.Size
}

func (f *parquetReader) Close() error {
	if f.debugMode {
		fmt.Printf("made %v calls cached calls %v\n", f.call, f.cached)
	}
	return nil
}

func (f *parquetReader) Seek(offset int64, whence int) (int64, error) {
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

func newParquetReader(ctx context.Context, bucket *blob.Bucket, obj *blob.ListObject) *parquetReader {
	return &parquetReader{
		ctx:       ctx,
		bucket:    bucket,
		obj:       obj,
		debugMode: true,
		buffer:    make([]byte, 1024*1024), // 1MB buffer
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
	reader := newParquetReader(ctx, bucket, obj)

	props := parquet.NewReaderProperties(mem)
	props.BufferedStreamEnabled = true

	pf, err := file.NewParquetReader(reader, file.WithReadProps(props))
	if err != nil {
		return err
	}
	defer pf.Close()

	// not 100% sure what is optimum BatchSize
	// from the code comments it seems like number of consecutive items in a column fetched in one shot across loading multiple row groups if required
	// since we have already enabled BufferedStreamEnabled for reading parquet, keeping it low shouldn't make extra network calls
	// whereas keeping it high can potentially load multiple groups(make multiple network calls)
	// keeping it 1 for simplicty
	arrowReadProperties := pqarrow.ArrowReadProperties{BatchSize: batchSize, Parallel: false}
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
