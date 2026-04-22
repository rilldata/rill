package static

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync/atomic"

	"github.com/c2h5oh/datasize"
	"github.com/mitchellh/mapstructure"
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
	Spec    *StaticSpec
	db      database.DB
	logger  *zap.Logger
	nextIdx atomic.Int64
}

var _ provisioner.Provisioner = (*StaticProvisioner)(nil)

func NewStatic(spec []byte, db database.DB, logger *zap.Logger) (provisioner.Provisioner, error) {
	sps := &StaticSpec{}
	err := json.Unmarshal(spec, sps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provisioner spec: %w", err)
	}

	p := &StaticProvisioner{
		Spec:   sps,
		db:     db,
		logger: logger,
	}

	// Initialize the round-robin index to a random value.
	p.nextIdx.Store(rand.Int63n(1000)) // nolint:gosec // We don't need secure random numbers

	return p, nil
}

func (p *StaticProvisioner) Type() string {
	return "static"
}

func (p *StaticProvisioner) Supports(rt provisioner.ResourceType) bool {
	return rt == provisioner.ResourceTypeRuntime
}

func (p *StaticProvisioner) Close() error {
	return nil
}

func (p *StaticProvisioner) Provision(ctx context.Context, r *provisioner.Resource, opts *provisioner.ResourceOptions) (*provisioner.Resource, error) {
	// Parse args
	args, err := provisioner.NewRuntimeArgs(opts.Args)
	if err != nil {
		return nil, err
	}

	// Parse state (if it's an update)
	state, err := newRuntimeState(r.State)
	if err != nil {
		return nil, err
	}

	// Exit early if the resource has already been provisioned.
	if state.Slots != 0 {
		if args.Slots != state.Slots {
			p.logger.Warn("static provisioner cannot update the slots assignment", observability.ZapCtx(ctx))
		}
		return r, nil
	}

	// Get slots currently used
	stats, err := p.db.ResolveStaticRuntimeSlotsUsed(ctx)
	if err != nil {
		return nil, err
	}
	hostToSlotsUsed := make(map[string]int, len(stats))
	for _, stat := range stats {
		hostToSlotsUsed[stat.Host] = stat.Slots
	}

	// Find runtimes with available capacity
	targets := make([]*StaticRuntimeSpec, 0)
	for _, candidate := range p.Spec.Runtimes {
		if hostToSlotsUsed[candidate.Host]+args.Slots <= candidate.Slots {
			targets = append(targets, candidate)
		}
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("no runtimes found with sufficient available slots")
	}

	// Select an eligible runtime using an approximately round-robin strategy
	idx := int((p.nextIdx.Add(1) - 1)) % len(targets)
	target := targets[idx]

	// Track slots used
	err = p.db.UpsertStaticRuntimeAssignment(ctx, r.ID, target.Host, args.Slots)
	if err != nil {
		return nil, err
	}

	// Build resource
	cfg := &provisioner.RuntimeConfig{
		Host:         target.Host,
		Audience:     target.Audience,
		CPU:          1 * args.Slots,
		MemoryGB:     4 * args.Slots,
		StorageBytes: int64(args.Slots) * 40 * int64(datasize.GB),
	}
	state = &runtimeState{
		Slots:   args.Slots,
		Version: opts.RillVersion,
	}
	return &provisioner.Resource{
		ID:     r.ID,
		Type:   r.Type,
		State:  state.AsMap(),
		Config: cfg.AsMap(),
	}, nil
}

func (p *StaticProvisioner) Deprovision(ctx context.Context, r *provisioner.Resource) error {
	// Check it's a runtime resource
	if r.Type != provisioner.ResourceTypeRuntime {
		return fmt.Errorf("unexpected resource type %q", r.Type)
	}

	// Remove the assignment
	return p.db.DeleteStaticRuntimeAssignment(ctx, r.ID)
}

func (p *StaticProvisioner) AwaitReady(ctx context.Context, r *provisioner.Resource) error {
	// No-op
	return nil
}

func (p *StaticProvisioner) Check(ctx context.Context) error {
	slotsUsedByRuntime, err := p.db.ResolveStaticRuntimeSlotsUsed(ctx)
	if err != nil {
		return err
	}

	var slotsTotal, slotsUsed int
	minPctUsed := 1.0

	for _, runtime := range p.Spec.Runtimes {
		slotsTotal += runtime.Slots
		for _, status := range slotsUsedByRuntime {
			if runtime.Host == status.Host {
				slotsUsed += status.Slots
				pctUsed := float64(status.Slots) / float64(runtime.Slots)
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

func (p *StaticProvisioner) CheckResource(ctx context.Context, r *provisioner.Resource, opts *provisioner.ResourceOptions) (*provisioner.Resource, error) {
	state, err := newRuntimeState(r.State)
	if err != nil {
		return nil, err
	}

	if state.Version != opts.RillVersion {
		// TODO: Instead of always updating the version, we should poll the runtime to check its current version.
		state.Version = opts.RillVersion
	}

	return &provisioner.Resource{
		ID:     r.ID,
		Type:   r.Type,
		State:  state.AsMap(),
		Config: r.Config,
	}, nil
}

// runtimeState describes the static provisioner's state for a provisioned runtime resource.
type runtimeState struct {
	Slots   int    `mapstructure:"slots"`
	Version string `mapstructure:"version"`
}

func newRuntimeState(state map[string]any) (*runtimeState, error) {
	res := &runtimeState{}
	err := mapstructure.Decode(state, res)
	if err != nil {
		return nil, fmt.Errorf("failed to parse static runtime state: %w", err)
	}
	return res, nil
}

func (r *runtimeState) AsMap() map[string]any {
	res := make(map[string]any)
	err := mapstructure.Decode(r, &res)
	if err != nil {
		panic(err)
	}
	return res
}
