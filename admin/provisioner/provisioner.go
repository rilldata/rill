package provisioner

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/c2h5oh/datasize"
	"github.com/rilldata/rill/admin/database"
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

type StaticSpec struct {
	Runtimes []*StaticRuntimeSpec `json:"runtimes"`
}

type StaticRuntimeSpec struct {
	Host     string `json:"host"`
	Region   string `json:"region"`
	Slots    int    `json:"slots"`
	DataDir  string `json:"data_dir"`
	Audience string `json:"audience_url"`
}

type StaticProvisioner struct {
	Spec *StaticSpec
	db   database.DB
}

func NewStatic(spec string, db database.DB) (*StaticProvisioner, error) {
	sps := &StaticSpec{}
	err := json.Unmarshal([]byte(spec), sps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provisioner spec: %w", err)
	}

	return &StaticProvisioner{
		Spec: sps,
		db:   db,
	}, nil
}

func (p *StaticProvisioner) Provision(ctx context.Context, opts *ProvisionOptions) (*Allocation, error) {
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
	targets := make([]*StaticRuntimeSpec, 0)
	for _, candidate := range p.Spec.Runtimes {
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
