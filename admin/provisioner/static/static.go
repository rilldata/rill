package static

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/c2h5oh/datasize"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/provisioner"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

func init() {
	provisioner.Register("static", NewStatic)
}

type StaticSpec struct {
	Runtimes []*StaticRuntimeSpec `json:"runtimes"`
}

type StaticRuntimeSpec struct {
	Host     string `json:"host"`
	Slots    int    `json:"slots"`
	Audience string `json:"audience_url"`
}

type StaticProvisioner struct {
	Spec   *StaticSpec
	db     database.DB
	logger *zap.Logger
}

var _ provisioner.Provisioner = (*StaticProvisioner)(nil)

func NewStatic(spec []byte, db database.DB, logger *zap.Logger) (provisioner.Provisioner, error) {
	sps := &StaticSpec{}
	err := json.Unmarshal(spec, sps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provisioner spec: %w", err)
	}

	return &StaticProvisioner{
		Spec:   sps,
		db:     db,
		logger: logger,
	}, nil
}

func (p *StaticProvisioner) Type() string {
	return "static"
}

func (p *StaticProvisioner) Provision(ctx context.Context, opts *provisioner.ProvisionOptions) (*provisioner.Resource, error) {
	// Can only provision runtime resources
	if opts.Type != provisioner.ResourceTypeRuntime {
		return nil, provisioner.ErrResourceTypeNotSupported
	}

	// Parse args
	args, err := provisioner.NewRuntimeArgs(opts.Args)
	if err != nil {
		return nil, err
	}

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
		if hostToSlotsUsed[candidate.Host]+args.Slots <= candidate.Slots {
			targets = append(targets, candidate)
		}
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no runtimes found with sufficient available slots")
	}

	// nolint:gosec // We don't need cryptographically secure random numbers
	target := targets[rand.Intn(len(targets))]
	cfg := &provisioner.RuntimeConfig{
		Host:         target.Host,
		Audience:     target.Audience,
		CPU:          1 * args.Slots,
		MemoryGB:     4 * args.Slots,
		StorageBytes: int64(args.Slots) * 40 * int64(datasize.GB),
	}

	// Done
	return &provisioner.Resource{
		ID:     opts.ID,
		Type:   opts.Type,
		Config: cfg.AsMap(),
	}, nil
}

func (p *StaticProvisioner) Deprovision(ctx context.Context, r *provisioner.Resource) error {
	// No-op
	return nil
}

func (p *StaticProvisioner) AwaitReady(ctx context.Context, r *provisioner.Resource) error {
	// No-op
	return nil
}

func (p *StaticProvisioner) Check(ctx context.Context) error {
	slotsUsedByRuntime, err := p.db.ResolveRuntimeSlotsUsed(ctx)
	if err != nil {
		return err
	}

	var slotsTotal, slotsUsed int
	minPctUsed := 1.0

	for _, runtime := range p.Spec.Runtimes {
		slotsTotal += runtime.Slots
		for _, status := range slotsUsedByRuntime {
			if runtime.Host == status.RuntimeHost {
				slotsUsed += status.SlotsUsed
				pctUsed := float64(status.SlotsUsed) / float64(runtime.Slots)
				if pctUsed < minPctUsed {
					minPctUsed = pctUsed
				}
			}
		}
	}

	// Log info status
	p.logger.Info(`slots check: status`, zap.Int("runtimes", len(p.Spec.Runtimes)), zap.Int("slots_total", slotsTotal), zap.Int("slots_used", slotsUsed), zap.Float64("min_pct_used", minPctUsed), observability.ZapCtx(ctx))

	// Check there's at least 20% free slots
	if float64(slotsUsed)/float64(slotsTotal) >= 0.8 {
		p.logger.Warn(`slots check: +80% of all slots used`, zap.Int("slots_total", slotsTotal), zap.Int("slots_used", slotsUsed), zap.Float64("min_pct_used", minPctUsed), observability.ZapCtx(ctx))
	}

	// Check there's at least one runtime with at least 30% free slots
	if slotsUsed != 0 && minPctUsed >= 0.7 {
		p.logger.Warn(`slots check: +70% of slots used on every runtime`, zap.Int("slots_total", slotsTotal), zap.Int("slots_used", slotsUsed), zap.Float64("min_pct_used", minPctUsed), observability.ZapCtx(ctx))
	}

	return nil
}

func (p *StaticProvisioner) CheckResource(ctx context.Context, r *provisioner.Resource) error {
	// No-op
	return nil
}
