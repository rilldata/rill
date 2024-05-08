package blob

import (
	"context"
	"fmt"
	"os"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"github.com/apache/arrow/go/v14/parquet"
	"github.com/apache/arrow/go/v14/parquet/compress"
	"github.com/apache/arrow/go/v14/parquet/file"
	"github.com/apache/arrow/go/v14/parquet/pqarrow"
	"github.com/rilldata/rill/runtime/pkg/arrayutil"
	"github.com/rilldata/rill/runtime/pkg/container"
	"gocloud.dev/blob"
)

// number of rows of a column fetched in one call
// keeping it high seems to improve latency at the cost of accuracy in size of fetched data as per policy
const _batchSize = int64(1024)

// downloadParquet downloads partial file as per extractOption
func downloadParquet(ctx context.Context, bucket *blob.Bucket, obj *blob.ListObject, option *extractOption, fw *os.File) error {
	reader := NewBlobObjectReader(ctx, bucket, obj)
	mem := memory.NewCheckedAllocator(memory.DefaultAllocator)
	props := parquet.NewReaderProperties(mem)

	pf, err := file.NewParquetReader(reader, file.WithReadProps(props))
	if err != nil {
		return err
	}
	defer pf.Close()

	arrowReadProperties := pqarrow.ArrowReadProperties{BatchSize: _batchSize, Parallel: true}
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
		parquet.WithCompression(compress.Codecs.Snappy),
		parquet.WithAllocator(mem))
	return pqarrow.WriteTable(
		table,
		fw,
		table.NumRows(),
		wp,
		pqarrow.NewArrowWriterProperties(pqarrow.WithAllocator(mem)),
	)
}

// estimateRecords estimates the number of rows to fetch based on extract policy
// each arrow.Record will hold batchSize number of rows
func estimateRecords(ctx context.Context, reader *file.Reader, pqToArrowReader *pqarrow.FileReader, config *extractOption) ([]arrow.Record, error) {
	rowIndexes := arrayutil.RangeInt(0, reader.NumRowGroups(), config.strategy == ExtractPolicyStrategyTail)

	var (
		// row group indices that we need
		reqRowIndices []int
		cumSize       uint64
		rows          int64
	)
	for _, index := range rowIndexes {
		reqRowIndices = append(reqRowIndices, index)
		rowGroup := reader.RowGroup(index)
		// we fetch compressed size from fs instead of ByteSize(uncompressed size)
		rowGroupSize := rowGroup.MetaData().TotalCompressedSize()
		rowCount := rowGroup.NumRows()

		if cumSize+uint64(rowGroupSize) > config.limitInBytes {
			// taking entire rowgroup crosses allowed size
			perRowSize := uint64(rowGroupSize / rowCount)
			rows += int64((config.limitInBytes - cumSize) / perRowSize)
			break
		}
		cumSize += uint64(rowGroupSize)
		rows += rowCount
	}

	r, err := pqToArrowReader.GetRecordReader(ctx, nil, reqRowIndices)
	if err != nil {
		return nil, err
	}
	defer r.Release()

	// one record has batchsize rows
	numRecords := rows / _batchSize
	if numRecords == 0 {
		// if parquet file has less than batchSize rows or user selects less than batchSize rows
		numRecords = 1
	}

	c, err := containerForRecordLimiting(config.strategy, int(numRecords))
	if err != nil {
		return nil, err
	}

	for r.Next() && !c.Full() {
		rec := r.Record()
		rec.Retain()
		c.Add(rec)
	}
	return c.Items(), nil
}

func containerForRecordLimiting(strategy ExtractPolicyStrategy, limit int) (container.Container[arrow.Record], error) {
	switch strategy {
	case ExtractPolicyStrategyHead:
		return container.NewBounded[arrow.Record](limit)
	case ExtractPolicyStrategyTail:
		return container.NewFIFO(limit, func(rec arrow.Record) { rec.Release() })
	default:
		// No option selected - this should not be used for partial downloads though
		// in case of no extract policy we should be directly downloading the entire file
		return container.NewUnbounded[arrow.Record]()
	}
}
