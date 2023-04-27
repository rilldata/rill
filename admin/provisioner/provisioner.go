package provisioner

import (
	"context"
	"encoding/json"
	"fmt"

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
	spec *staticSpec
	db   database.DB
}

func NewStatic(spec string, db database.DB) (Provisioner, error) {
	sps := &staticSpec{}
	err := json.Unmarshal([]byte(spec), sps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provisioner spec: %w", err)
	}

	return &staticProvisioner{
		spec: sps,
		db:   db,
	}, nil
}

func (p *staticProvisioner) Provision(ctx context.Context, opts *ProvisionOptions) (*Allocation, error) {
	// Get slots currently used
	stats, err := p.db.ResolveRuntimeSlotsUsed(ctx)
	if err != nil {
		return nil, err
	}

	// Find runtime with available capacity
	var target *staticRuntime
	for _, candidate := range p.spec.Runtimes {
		if opts.Region != "" && opts.Region != candidate.Region {
			continue
		}

		available := true
		for _, stat := range stats {
			if stat.RuntimeHost == candidate.Host && stat.SlotsUsed+opts.Slots > candidate.Slots {
				available = false
				break
			}
		}

		if available {
			target = candidate
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("no runtimes found with sufficient available slots")
	}

	return &Allocation{
		Host:         target.Host,
		Audience:     target.Audience,
		DataDir:      target.DataDir,
		CPU:          1 * opts.Slots,
		MemoryGB:     2 * opts.Slots,
		StorageBytes: int64(opts.Slots) * 5 * int64(datasize.GB),
	}, nil
}
