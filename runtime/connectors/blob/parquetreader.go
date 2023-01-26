package blob

import (
	"container/list"
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

// keeping it high seems to improve latency at the cost of accuracy in size of fetched data as per policy
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
		return n, err
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

// todo :: see if recordContainer and container can be implemented in a single generic way
// recordContainer keeps items as per extract config
type recordContainer struct {
	config *ExtractConfig
	items  *list.List
}

func (c *recordContainer) Add(record arrow.Record) bool {
	if c.IsFull() {
		return false
	}

	switch c.config.Strategy {
	case TAIL:
		// keep latest item at front
		c.items.PushFront(record)
		if c.items.Len() > int(c.config.Size) {
			// remove oldest item
			record := c.items.Remove(c.items.Back()).(arrow.Record)
			record.Release()
		}
	case HEAD:
		c.items.PushBack(record)
	default:
		c.items.PushBack(record)
	}
	return true
}

func (c *recordContainer) IsFull() bool {
	switch c.config.Strategy {
	case TAIL:
		return false
	case HEAD:
		return c.items.Len() >= int(c.config.Size)
	default:
		return false
	}
}

func (c *recordContainer) Items() []arrow.Record {
	result := make([]arrow.Record, c.items.Len())

	for front, i := c.items.Front(), 0; front != nil; front, i = c.items.Front(), i+1 {
		result[i] = c.items.Remove(front).(arrow.Record)
	}
	return result
}

func newRecordContainer(config *ExtractConfig) *recordContainer {
	return &recordContainer{config: config, items: list.New().Init()}
}

func estimateRecords(ctx context.Context, reader *file.Reader, pqToArrowReader *pqarrow.FileReader, config ExtractConfig) ([]arrow.Record, error) {
	rowIndexes := getArray(reader.NumRowGroups(), config.Strategy == TAIL)

	// row group indices that we need
	reqRowIndices := make([]int, 0)
	var cumSize, rows int64
	for _, index := range rowIndexes {
		reqRowIndices = append(reqRowIndices, index)
		rowGroup := reader.RowGroup(index)
		rowGroupSize := rowGroup.ByteSize()
		rowCount := rowGroup.NumRows()

		if cumSize+rowGroupSize > config.Size {
			// taking entire rowgroup crosses allowed size
			perRowSize := rowGroupSize / rowCount
			rows += (config.Size - cumSize) / perRowSize
			break
		}
		rows += rowCount
	}

	r, err := pqToArrowReader.GetRecordReader(ctx, getArray(reader.RowGroup(0).NumColumns(), false), reqRowIndices)
	if err != nil {
		return nil, err
	}
	defer r.Release()

	// one record has batchsize rows
	numRecords := rows / batchSize
	if numRecords == 0 {
		// if parquet file has less than batchSize rows or user selects less than batchSize rows
		numRecords = 1
	}

	container := newRecordContainer(&ExtractConfig{Strategy: config.Strategy, Size: numRecords})
	for r.Next() && !container.IsFull() {
		rec := r.Record()
		rec.Retain()
		container.Add(rec)
	}
	return container.Items(), nil
}

func DownloadParquet(ctx context.Context, bucket *blob.Bucket, obj *blob.ListObject, option ExtractConfig, fw *os.File) error {
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
	// reader to convert parquet objects to arrow objects
	fileReader, err := pqarrow.NewFileReader(pf, arrowReadProperties, mem)
	if err != nil {
		return err
	}

	numRowGroups := pf.NumRowGroups()
	if numRowGroups == 0 {
		return fmt.Errorf("invalid parquet")
	}

	records, err := estimateRecords(ctx, pf, fileReader, option)
	if err != nil {
		return err
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
