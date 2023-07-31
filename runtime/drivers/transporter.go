package drivers

import (
	"context"
	"math"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Transporter implements logic for moving data between two connectors
// (the actual connector objects are provided in AsTransporter)
type Transporter interface {
	Transfer(ctx context.Context, source Source, sink Sink, t *TransferOpts, p Progress) error
}

type TransferOpts struct {
	IteratorBatch int
	LimitInBytes  int64
}

func NewTransferOpts(opts ...TransferOption) *TransferOpts {
	t := &TransferOpts{
		IteratorBatch: _iteratorBatch,
		LimitInBytes:  math.MaxInt64,
	}

	for _, opt := range opts {
		opt(t)
	}
	return t
}

type TransferOption func(*TransferOpts)

func WithIteratorBatch(b int) TransferOption {
	return func(t *TransferOpts) {
		t.IteratorBatch = b
	}
}

func WithLimitInBytes(limit int64) TransferOption {
	return func(t *TransferOpts) {
		t.LimitInBytes = limit
	}
}

// A Source is expected to only return ok=true for one of the source types.
// The caller will know which type based on the connector type.
type Source interface {
	BucketSource() (*BucketSource, bool)
	DatabaseSource() (*DatabaseSource, bool)
	FileSource() (*FileSource, bool)
}

// A Sink is expected to only return ok=true for one of the sink types.
// The caller will know which type based on the connector type.
type Sink interface {
	BucketSink() (*BucketSink, bool)
	DatabaseSink() (*DatabaseSink, bool)
}

type BucketSource struct {
	ExtractPolicy *runtimev1.Source_ExtractPolicy
	Properties    map[string]any
}

var _ Source = &BucketSource{}

func (b *BucketSource) BucketSource() (*BucketSource, bool) {
	return b, true
}

func (b *BucketSource) DatabaseSource() (*DatabaseSource, bool) {
	return nil, false
}

func (b *BucketSource) FileSource() (*FileSource, bool) {
	return nil, false
}

type BucketSink struct {
	Path string
	// Format FileFormat
	// NOTE: In future, may add file name and output partitioning config here
}

var _ Sink = &BucketSink{}

func (b *BucketSink) BucketSink() (*BucketSink, bool) {
	return b, true
}

func (b *BucketSink) DatabaseSink() (*DatabaseSink, bool) {
	return nil, false
}

type DatabaseSource struct {
	// Pass only SQL OR Table
	SQL      string
	Table    string
	Database string
	Limit    int
	Props    map[string]any
}

var _ Source = &DatabaseSource{}

func (d *DatabaseSource) BucketSource() (*BucketSource, bool) {
	return nil, false
}

func (d *DatabaseSource) DatabaseSource() (*DatabaseSource, bool) {
	return d, true
}

func (d *DatabaseSource) FileSource() (*FileSource, bool) {
	return nil, false
}

type DatabaseSink struct {
	Table  string
	Append bool
}

var _ Sink = &DatabaseSink{}

func (d *DatabaseSink) BucketSink() (*BucketSink, bool) {
	return nil, false
}

func (d *DatabaseSink) DatabaseSink() (*DatabaseSink, bool) {
	return d, true
}

type FileSource struct {
	Name       string
	Properties map[string]any
}

var _ Source = &FileSource{}

func (f *FileSource) BucketSource() (*BucketSource, bool) {
	return nil, false
}

func (f *FileSource) DatabaseSource() (*DatabaseSource, bool) {
	return nil, false
}

func (f *FileSource) FileSource() (*FileSource, bool) {
	return f, true
}

// Progress is an interface for communicating progress info
type Progress interface {
	Target(val int64, unit ProgressUnit)
	// Observe is used by caller to provide incremental updates
	Observe(val int64, unit ProgressUnit)
}

type NoOpProgress struct{}

func (n NoOpProgress) Target(val int64, unit ProgressUnit)  {}
func (n NoOpProgress) Observe(val int64, unit ProgressUnit) {}

var _ Progress = NoOpProgress{}

type ProgressUnit int

const (
	ProgressUnitByte ProgressUnit = iota
	ProgressUnitFile
	ProgressUnitRecord
)
