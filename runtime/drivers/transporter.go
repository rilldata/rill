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
	AcquireConnector func(string) (Handle, func(), error)
}

type ProgressUnit int

const (
	ProgressUnitByte ProgressUnit = iota
	ProgressUnitFile
	ProgressUnitRecord
)
