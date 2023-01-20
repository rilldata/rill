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
	"github.com/apache/arrow/go/v11/parquet/file"
	"github.com/apache/arrow/go/v11/parquet/pqarrow"
	"gocloud.dev/blob"
)

type parquetReader struct {
	ctx    context.Context
	bucket *blob.Bucket
	index  int64
	obj    *blob.ListObject
	call   *int32
}

// todo :: add buffer for caching
func (f *parquetReader) ReadAt(p []byte, off int64) (n int, err error) {
	fmt.Printf("reading %v bytes at offset %v\n", len(p), off)

	reader, err := f.bucket.NewRangeReader(f.ctx, f.obj.Key, off, int64(len(p)), nil)
	// reader, err := f.bucket.NewReader(f.ctx, f.fileName, nil)
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	read, readErr := reader.Read(p)
	if readErr != nil {
		return n, readErr
	}

	atomic.AddInt32(f.call, 1)
	return read, nil
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
	fmt.Printf("made %v calls\n", *f.call)
	return nil
	// return f.bucket.Close()
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

func NewParquetReader(ctx context.Context, bucket *blob.Bucket, obj *blob.ListObject) *parquetReader {
	return &parquetReader{
		ctx:    ctx,
		bucket: bucket,
		index:  0,
		obj:    obj,
		call:   new(int32),
	}
}

func panicIfError(err error) {
	if err != nil {
		fmt.Printf("err %v\n", err)
		panic(err)
	}
}

func reverse(s []int) {
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
		reverse(result)
	}
	return result
}

func estimate(reader *file.Reader, option ExtractOptions) ([]int, int64) {
	rowIndexes := getArray(reader.NumRowGroups(), option.Strategy == TAIL)

	result := make([]int, 0)
	var cumSize, rows int64
	for _, index := range rowIndexes {
		result = append(result, index)
		rowGroup := reader.RowGroup(index)
		rowGroupSize := rowGroup.ByteSize()
		rowCount := rowGroup.NumRows()

		if cumSize+rowGroupSize > option.Size {
			// taking entire crosses allowed size
			perRowSize := rowGroupSize / rowCount
			rows += int64((option.Size - cumSize) / perRowSize)
			return result, rows
		}
		rows += rowCount
	}
	return result, rows
}

func Download(ctx context.Context, bucket *blob.Bucket, obj *blob.ListObject, option ExtractOptions) (string, error) {
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	reader := NewParquetReader(ctx, bucket, obj)

	props := parquet.NewReaderProperties(mem)
	props.BufferedStreamEnabled = true
	props.BufferSize = 1024 * 1024

	pf, err := file.NewParquetReader(reader, file.WithReadProps(props))
	if err != nil {
		return "", err
	}
	defer pf.Close()

	// not 100% sure what is optimum BatchSize
	// from the code comments it seems like number of consecutive items in a column fetched in one shot across loading multiple row groups if required
	// since we have already enabled BufferedStreamEnabled for reading parquet keeping it low shouldn't make extra network calls
	// whereas keeping it high can potentially load multiple groups(make multiple network calls)
	// keeping it 1 for simplicty
	arrowReadProperties := pqarrow.ArrowReadProperties{BatchSize: 1, Parallel: true}
	fileReader, err := pqarrow.NewFileReader(pf, arrowReadProperties, mem)
	if err != nil {
		return "", err
	}

	numRowGroups := pf.NumRowGroups()
	if numRowGroups == 0 {
		panicIfError(fmt.Errorf("invalid parquet"))
	}

	rowIndices, rowLimit := estimate(pf, option)

	r, err := fileReader.GetRecordReader(ctx, getArray(pf.RowGroup(0).NumColumns(), false), rowIndices)
	if err == nil {
		return "", err
	}

	records := make([]arrow.Record, rowLimit)
	for i := int64(0); i < rowLimit; i++ {
		rec, err := r.Read()
		if err != nil {
			return "", err
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
		return "", err
	}

	table := array.NewTableFromRecords(schema, records)
	defer table.Release()

	// println(table.NumRows())
	fileName := "out.parquet"
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}

	defer file.Close()

	if err := pqarrow.WriteTable(table, file, table.NumRows(), nil, pqarrow.NewArrowWriterProperties(pqarrow.WithAllocator(mem))); err != nil {
		return "", err
	}
	return fileName, nil
}
