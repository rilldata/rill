package bigquery

import (
	"bytes"
	"errors"
	"sync/atomic"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/apache/arrow/go/v13/arrow"
	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/apache/arrow/go/v13/arrow/ipc"
	"github.com/apache/arrow/go/v13/arrow/memory"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

func (iter *fileIterator) AsArrowRecordReader() (array.RecordReader, error) {
	arrowIt, err := iter.bqIter.ArrowIterator()
	if err != nil {
		return nil, err
	}

	allocator := memory.DefaultAllocator
	buf := bytes.NewBuffer(arrowIt.SerializedArrowSchema())
	rdr, err := ipc.NewReader(buf, ipc.WithAllocator(allocator))
	if err != nil {
		return nil, err
	}
	defer rdr.Release()

	rec := &arrowRecordReader{
		bqIter:      arrowIt,
		arrowSchema: rdr.Schema(),
		refCount:    1,
		allocator:   allocator,
		logger:      iter.logger,
		records:     make([]arrow.Record, 0),
	}

	return rec, rec.err
}

// some impl details are copied from array.simpleRecords
type arrowRecordReader struct {
	bqIter      bigquery.ArrowIterator
	records     []arrow.Record
	cur         arrow.Record
	arrowSchema *arrow.Schema
	refCount    int64
	err         error
	logger      *zap.Logger
	allocator   memory.Allocator

	apinext time.Duration
	ipcread time.Duration
}

// Retain increases the reference count by 1.
// Retain may be called simultaneously from multiple goroutines.
func (rs *arrowRecordReader) Retain() {
	atomic.AddInt64(&rs.refCount, 1)
}

// Release decreases the reference count by 1.
// When the reference count goes to zero, the memory is freed.
// Release may be called simultaneously from multiple goroutines.
func (rs *arrowRecordReader) Release() {
	rs.logger.Info("next api call took", zap.Float64("seconds", rs.apinext.Seconds()))
	rs.logger.Info("next ipc read took", zap.Float64("seconds", rs.ipcread.Seconds()))
	if atomic.LoadInt64(&rs.refCount) <= 0 {
		return
	}

	if atomic.AddInt64(&rs.refCount, -1) == 0 {
		if rs.cur != nil {
			rs.cur.Release()
		}
		for _, rec := range rs.records {
			rec.Release()
		}
		rs.records = nil
	}
}

// Schema returns the underlying arrow schema
func (rs *arrowRecordReader) Schema() *arrow.Schema {
	return rs.arrowSchema
}

// Record returns the current record. Call Next before consuming another record.
func (rs *arrowRecordReader) Record() arrow.Record {
	return rs.cur
}

// Next returns true if another record can be produced
func (rs *arrowRecordReader) Next() bool {
	if rs.err != nil {
		return false
	}

	if len(rs.records) == 0 {
		tz := time.Now()
		next, err := rs.bqIter.Next()
		if err != nil {
			rs.err = err
			return false
		}
		rs.apinext += time.Since(tz)

		rs.records, rs.err = rs.nextArrowRecords(next)
		if rs.err != nil {
			return false
		}
	}
	if rs.cur != nil {
		rs.cur.Release()
	}
	rs.cur = rs.records[0]
	rs.records = rs.records[1:]
	return true
}

func (rs *arrowRecordReader) Err() error {
	if errors.Is(rs.err, iterator.Done) {
		return drivers.ErrIteratorDone
	}
	return rs.err
}

func (rs *arrowRecordReader) nextArrowRecords(r *bigquery.ArrowRecordBatch) ([]arrow.Record, error) {
	t := time.Now()
	defer func() {
		rs.ipcread += time.Since(t)
	}()

	buf := bytes.NewBuffer(rs.bqIter.SerializedArrowSchema())
	buf.Write(r.Data)
	rdr, err := ipc.NewReader(buf, ipc.WithSchema(rs.arrowSchema), ipc.WithAllocator(rs.allocator))
	if err != nil {
		return nil, err
	}
	defer rdr.Release()
	records := make([]arrow.Record, 0)
	for rdr.Next() {
		rec := rdr.Record()
		rec.Retain()
		records = append(records, rec)
	}
	return records, nil
}
