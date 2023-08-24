package bigquery

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/apache/arrow/go/v13/arrow"
	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/apache/arrow/go/v13/arrow/ipc"
	"github.com/apache/arrow/go/v13/arrow/memory/mallocator"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/api/iterator"
)

func AsArrowRecordReader(i drivers.RowIterator) (array.RecordReader, error) {
	t := time.Now()
	defer func() {
		log.Default().Printf("fetching arrow recorder took %v", time.Since(t).Seconds())
	}()
	iter, ok := i.(*rowIterator)
	if !ok || iter.bqIter.ArrowIterator == nil {
		return nil, fmt.Errorf("not using storage API")
	}

	tz := time.Now()
	ser, err := iter.bqIter.ArrowIterator.Next()
	if err != nil {
		return nil, err
	}
	duration := time.Since(tz)

	arrowSerializedSchema := iter.bqIter.ArrowIterator.Decoder.RawArrowSchema
	buf := bytes.NewBuffer(arrowSerializedSchema)
	rdr, err := ipc.NewReader(buf, ipc.WithAllocator(mallocator.NewMallocator()))
	if err != nil {
		return nil, err
	}
	defer rdr.Release()
	rec := &arrowRecordReader{
		r:           iter,
		arrowSchema: rdr.Schema(),
		refCount:    1,
		records:     make([]arrow.Record, 0),
		apinext:     duration,
	}

	rec.records, rec.err = rec.nextArrowRecords(ser)
	return rec, rec.err
}

type arrowRecordReader struct {
	r           *rowIterator
	records     []arrow.Record
	cur         arrow.Record
	arrowSchema *arrow.Schema
	refCount    int64
	err         error
	t           time.Duration
	apinext     time.Duration
	ipcread     time.Duration
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
	log.Default().Printf("next took %v\n", rs.t.Seconds())
	log.Default().Printf("next api took %v\n", rs.apinext.Seconds())
	log.Default().Printf("next ipc read took %v\n", rs.ipcread.Seconds())
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

func (rs *arrowRecordReader) Schema() *arrow.Schema {
	return rs.arrowSchema
}

func (rs *arrowRecordReader) Record() arrow.Record {
	return rs.cur
}

func (rs *arrowRecordReader) Next() bool {
	tw := time.Now()
	defer func() {
		rs.t += time.Since(tw)
	}()
	if rs.err != nil {
		return false
	}

	if len(rs.records) == 0 {
		tz := time.Now()
		next, err := rs.r.bqIter.ArrowIterator.Next()
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
		return nil
	}
	return rs.err
}

func (rs *arrowRecordReader) nextArrowRecords(serializedRecord []byte) ([]arrow.Record, error) {
	t := time.Now()
	defer func() {
		rs.ipcread += time.Since(t)
	}()

	buf := bytes.NewBuffer(rs.r.bqIter.ArrowIterator.Decoder.RawArrowSchema)
	buf.Write(serializedRecord)
	rdr, err := ipc.NewReader(buf, ipc.WithSchema(rs.arrowSchema), ipc.WithAllocator(mallocator.NewMallocator()))
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
