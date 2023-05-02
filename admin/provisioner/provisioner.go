package provisioner

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/c2h5oh/datasize"
	"github.com/rilldata/rill/admin/database"
	"go.uber.org/zap"
)

type Allocation struct {
	Host         string
	Audience     string
	DataDir      string
	CPU          int
	MemoryGB     int
	StorageBytes int64
}

type ProvisionOptions struct {
	OLAPDriver string
	Slots      int
	Region     string
}

type Provisioner interface {
	Provision(ctx context.Context, opts *ProvisionOptions) (*Allocation, error)
}

type staticSpec struct {
	Runtimes []*staticRuntime `json:"runtimes"`
}

type staticRuntime struct {
	Host     string `json:"host"`
	Region   string `json:"region"`
	Slots    int    `json:"slots"`
	DataDir  string `json:"data_dir"`
	Audience string `json:"audience_url"`
}

type staticProvisioner struct {
	spec   *staticSpec
	db     database.DB
	logger *zap.Logger
}

func NewStatic(spec string, db database.DB, logger *zap.Logger) (Provisioner, error) {
	sps := &staticSpec{}
	err := json.Unmarshal([]byte(spec), sps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provisioner spec: %w", err)
	}

	return &staticProvisioner{
		spec:   sps,
		db:     db,
		logger: logger,
	}, nil
}

func (p *staticProvisioner) Provision(ctx context.Context, opts *ProvisionOptions) (*Allocation, error) {
	// Get slots currently used
	stats, err := p.db.ResolveRuntimeSlotsUsed(ctx)
	if err != nil {
		return nil, err
	}

	hostToSlotsUsed := make(map[string]int, len(stats))
	for _, stat := range stats {
		hostToSlotsUsed[stat.RuntimeHost] = stat.SlotsUsed
	}

	// Find runtime with available capacity
	targets := make([]*staticRuntime, 0)
	for _, candidate := range p.spec.Runtimes {
		if opts.Region != "" && opts.Region != candidate.Region {
			continue
		}

		if hostToSlotsUsed[candidate.Host]+opts.Slots <= candidate.Slots {
			targets = append(targets, candidate)
		}
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no runtimes found with sufficient available slots")
	}

	// nolint:gosec // We don't need cryptographically secure random numbers
	target := targets[rand.Intn(len(targets))]
	return &Allocation{
		Host:         target.Host,
		Audience:     target.Audience,
		DataDir:      target.DataDir,
		CPU:          1 * opts.Slots,
		MemoryGB:     2 * opts.Slots,
		StorageBytes: int64(opts.Slots) * 5 * int64(datasize.GB),
	}, nil
}
