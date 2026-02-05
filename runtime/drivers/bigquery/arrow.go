package bigquery

import (
	"bytes"
	"context"
	"errors"
	"sync/atomic"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/ipc"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

func (f *fileIterator) AsArrowRecordReader(ctx context.Context) (array.RecordReader, error) {
	arrowIt, err := f.bqIter.ArrowIterator()
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
		refCount:    atomic.Int64{},
		allocator:   allocator,
		logger:      f.logger,
		records:     make([]arrow.RecordBatch, 0),
		ctx:         ctx,
		buf:         &bytes.Buffer{},
	}
	rec.refCount.Store(1)
	return rec, nil
}

// some impl details are copied from array.simpleRecords
type arrowRecordReader struct {
	bqIter      bigquery.ArrowIterator
	records     []arrow.RecordBatch
	cur         arrow.RecordBatch
	arrowSchema *arrow.Schema
	refCount    atomic.Int64
	err         error
	logger      *zap.Logger
	allocator   memory.Allocator

	apinext time.Duration
	ipcread time.Duration

	ctx context.Context
	buf *bytes.Buffer
}

// Retain increases the reference count by 1.
// Retain may be called simultaneously from multiple goroutines.
func (rs *arrowRecordReader) Retain() {
	rs.refCount.Add(1)
}

// Release decreases the reference count by 1.
// When the reference count goes to zero, the memory is freed.
// Release may be called simultaneously from multiple goroutines.
func (rs *arrowRecordReader) Release() {
	if rs.refCount.Load() <= 0 {
		rs.logger.Warn("too many releases", observability.ZapCtx(rs.ctx))
		return
	}

	if rs.refCount.Add(-1) == 0 {
		if rs.cur != nil {
			rs.cur.Release()
		}
		for _, rec := range rs.records {
			rec.Release()
		}
		rs.records = nil
	}
	rs.logger.Debug("next call took", zap.Float64("apinext_seconds", rs.apinext.Seconds()), zap.Float64("ipcread_seconds", rs.ipcread.Seconds()), observability.ZapCtx(rs.ctx))
}

// Schema returns the underlying arrow schema
func (rs *arrowRecordReader) Schema() *arrow.Schema {
	return rs.arrowSchema
}

// Record returns the current record. Call Next before consuming another record.
func (rs *arrowRecordReader) Record() arrow.RecordBatch {
	return rs.RecordBatch()
}

func (rs *arrowRecordReader) RecordBatch() arrow.RecordBatch {
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
		rs.cur = nil
	}
	rs.cur = rs.records[0]
	rs.records = rs.records[1:]
	return true
}

func (rs *arrowRecordReader) Err() error {
	if errors.Is(rs.err, iterator.Done) {
		return nil
	}
	return rs.err
}

func (rs *arrowRecordReader) nextArrowRecords(r *bigquery.ArrowRecordBatch) ([]arrow.RecordBatch, error) {
	t := time.Now()
	defer func() {
		rs.ipcread += time.Since(t)
	}()

	rs.buf.Reset()
	rs.buf.Write(rs.bqIter.SerializedArrowSchema())
	rs.buf.Write(r.Data)
	rdr, err := ipc.NewReader(rs.buf, ipc.WithSchema(rs.arrowSchema), ipc.WithAllocator(rs.allocator))
	if err != nil {
		return nil, err
	}
	defer rdr.Release()
	records := make([]arrow.RecordBatch, 0)
	for rdr.Next() {
		rec := rdr.RecordBatch()
		rec.Retain()
		records = append(records, rec)
	}
	return records, rdr.Err()
}
