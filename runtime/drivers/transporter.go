package drivers

import (
	"context"
	"math"
)

// Transporter implements logic for moving data between two connectors
// (the actual connector objects are provided in AsTransporter)
type Transporter interface {
	Transfer(ctx context.Context, source map[string]any, sink map[string]any, t *TransferOpts, p Progress) error
}

type TransferOpts struct {
	IteratorBatch            int
	IteratorBatchSizeInBytes int64
	LimitInBytes             int64
}

func NewTransferOpts(opts ...TransferOption) *TransferOpts {
	t := &TransferOpts{
		IteratorBatch:            _iteratorBatch,
		LimitInBytes:             math.MaxInt64,
		IteratorBatchSizeInBytes: _iteratorBatchSizeInBytes,
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

func WithIteratorBatchSizeInBytes(b int64) TransferOption {
	return func(t *TransferOpts) {
		t.IteratorBatchSizeInBytes = b
	}
}

func WithLimitInBytes(limit int64) TransferOption {
	return func(t *TransferOpts) {
		t.LimitInBytes = limit
	}
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
