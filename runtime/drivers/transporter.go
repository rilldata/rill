package drivers

import (
	"context"
)

// Transporter implements logic for moving data between two connectors
// (the actual connector objects are provided in AsTransporter)
type Transporter interface {
	Transfer(ctx context.Context, srcProps, sinkProps map[string]any, opts *TransferOptions) error
}

// TransferOptions provide execution context for Transporter.Transfer
type TransferOptions struct {
	AllowHostAccess  bool
	RepoRoot         string
	LimitInBytes     int64
	Progress         Progress
	AcquireConnector func(string) (Handle, func(), error)
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
