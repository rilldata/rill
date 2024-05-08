package reconcilers

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const _defaultIngestTimeout = 60 * time.Minute

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindSource, newSourceReconciler)
}

type SourceReconciler struct {
	C *runtime.Controller
}

func newSourceReconciler(c *runtime.Controller) runtime.Reconciler {
	return &SourceReconciler{C: c}
}

func (r *SourceReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *SourceReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetSource()
	b := to.GetSource()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *SourceReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetSource()
	b := to.GetSource()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *SourceReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetSource().State = &runtimev1.SourceState{}
	return nil
}

func (r *SourceReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	src := self.GetSource()
	if src == nil {
		return runtime.ReconcileResult{Err: errors.New("not a source")}
	}

	// The table name to ingest into is derived from the resource name.
	// We only set src.State.Table after ingestion is complete.
	// The value of tableName and src.State.Table will differ until initial successful ingestion and when renamed.
	tableName := self.Meta.Name.Name

	// Handle deletion
	if self.Meta.DeletedOn != nil {
		olapDropTableIfExists(ctx, r.C, src.State.Connector, src.State.Table, false)
		olapDropTableIfExists(ctx, r.C, src.State.Connector, r.stagingTableName(tableName), false)
		return runtime.ReconcileResult{}
	}

	// Handle renames
	if self.Meta.RenamedFrom != nil {
		// Check if the table exists (it should, but might somehow have been corrupted)
		_, ok := olapTableInfo(ctx, r.C, src.State.Connector, src.State.Table)
		// NOTE: Not checking if it's a view because some backends will represent sources as views (like DuckDB with external table storage enabled).
		if ok {
			// Rename and update state
			err = olapForceRenameTable(ctx, r.C, src.State.Connector, src.State.Table, false, tableName)
			if err != nil {
				return runtime.ReconcileResult{Err: fmt.Errorf("failed to rename table: %w", err)}
			}
			src.State.Table = tableName
			err = r.C.UpdateState(ctx, self.Meta.Name, self)
			if err != nil {
				return runtime.ReconcileResult{Err: err}
			}
		}
		// Note: Not exiting early. It might need to be (re-)ingested, and we need to set the correct retrigger time based on the refresh schedule.
	}

	// Exit early if disabled
	if src.Spec.RefreshSchedule != nil && src.Spec.RefreshSchedule.Disable {
		return runtime.ReconcileResult{}
	}

	// Check refs - stop if any of them are invalid
	err = checkRefs(ctx, r.C, self.Meta.Refs)
	if err != nil {
		if !src.Spec.StageChanges && src.State.Table != "" {
			// Remove previously ingested table
			olapDropTableIfExists(ctx, r.C, src.State.Connector, src.State.Table, false)
			src.State.Connector = ""
			src.State.Table = ""
			src.State.SpecHash = ""
			src.State.RefreshedOn = nil
			err = r.C.UpdateState(ctx, self.Meta.Name, self)
			if err != nil {
				r.C.Logger.Error("refs check: failed to update state", zap.Any("error", err))
			}
		}
		return runtime.ReconcileResult{Err: err}
	}

	srcConfig, err := r.driversSource(ctx, self, src.Spec.Properties)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	// Use a hash of ingestion-related fields from the spec to determine if we need to re-ingest
	hash, err := r.ingestionSpecHash(src.Spec, srcConfig)
	if err != nil {
		return runtime.ReconcileResult{Err: fmt.Errorf("failed to compute hash: %w", err)}
	}

	// Compute next time to refresh based on the RefreshSchedule (if any)
	var refreshOn time.Time
	if src.State.RefreshedOn != nil {
		refreshOn, err = nextRefreshTime(src.State.RefreshedOn.AsTime(), src.Spec.RefreshSchedule)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Check if the table still exists (might have been corrupted/lost somehow)
	tableExists := false
	if src.State.Table != "" {
		_, ok := olapTableInfo(ctx, r.C, src.State.Connector, src.State.Table)
		tableExists = ok // NOTE: Not checking if it's a view because some backends will represent sources as views (like DuckDB with external table storage enabled)
	}

	// Decide if we should trigger a refresh
	trigger := src.Spec.Trigger                                             // If Trigger is set
	trigger = trigger || src.State.Table == ""                              // If table is missing
	trigger = trigger || src.State.RefreshedOn == nil                       // If never refreshed
	trigger = trigger || src.State.SpecHash != hash                         // If the spec has changed
	trigger = trigger || !tableExists                                       // If the table has disappeared
	trigger = trigger || !refreshOn.IsZero() && time.Now().After(refreshOn) // If the schedule says it's time

	// Exit early if no trigger
	if !trigger {
		return runtime.ReconcileResult{Retrigger: refreshOn}
	}

	// If the SinkConnector was changed, drop data in the old connector
	if src.State.Table != "" && src.State.Connector != src.Spec.SinkConnector {
		olapDropTableIfExists(ctx, r.C, src.State.Connector, src.State.Table, false)
		olapDropTableIfExists(ctx, r.C, src.State.Connector, r.stagingTableName(src.State.Table), false)
	}

	// Prepare for ingestion
	stagingTableName := tableName
	connector := src.Spec.SinkConnector
	if src.Spec.StageChanges {
		stagingTableName = r.stagingTableName(tableName)
	}

	// Should never happen, but if somehow the staging table was corrupted into a view, drop it
	if t, ok := olapTableInfo(ctx, r.C, connector, stagingTableName); ok && t.View {
		olapDropTableIfExists(ctx, r.C, connector, stagingTableName, t.View)
	}

	// Execute ingestion
	r.C.Logger.Info("Ingesting source data", zap.String("name", n.Name), zap.String("connector", connector))
	ingestErr := r.ingestSource(ctx, self, srcConfig, driversSink(stagingTableName))
	if ingestErr != nil {
		ingestErr = fmt.Errorf("failed to ingest source: %w", ingestErr)
	}

	if ingestErr == nil && src.Spec.StageChanges {
		// Rename staging table to main table
		err = olapForceRenameTable(ctx, r.C, connector, stagingTableName, false, tableName)
		if err != nil {
			return runtime.ReconcileResult{Err: fmt.Errorf("failed to rename staging table: %w", err)}
		}
	}

	// How we handle ingestErr depends on several things:
	// If ctx was cancelled, we cleanup and exit
	// If StageChanges is true, we retain the existing table, but still return the error.
	// If StageChanges is false, we clear the existing table and return the error.

	// ctx will only be cancelled in cases where the Controller guarantees a new call to Reconcile.
	// We just clean up temp tables and state, then return.
	cleanupCtx := ctx
	if ctx.Err() != nil {
		var cancel context.CancelFunc
		cleanupCtx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
	}

	// Update state
	update := false
	if ingestErr == nil {
		// Successful ingestion
		update = true
		src.State.Connector = connector
		src.State.Table = tableName
		src.State.SpecHash = hash
		src.State.RefreshedOn = timestamppb.Now()
	} else if src.Spec.StageChanges {
		// Failed ingestion to staging table
		olapDropTableIfExists(cleanupCtx, r.C, connector, stagingTableName, false)
	} else {
		// Failed ingestion to main table
		update = true
		olapDropTableIfExists(cleanupCtx, r.C, connector, tableName, false)
		src.State.Connector = ""
		src.State.Table = ""
		src.State.SpecHash = ""
		src.State.RefreshedOn = nil
	}
	if update {
		err = r.C.UpdateState(ctx, self.Meta.Name, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// See earlier note – essential cleanup is done, we can return now.
	if ctx.Err() != nil {
		return runtime.ReconcileResult{Err: ingestErr}
	}

	// Reset spec.Trigger
	if src.Spec.Trigger {
		err := r.setTriggerFalse(ctx, n)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Compute next refresh time
	refreshOn, err = nextRefreshTime(time.Now(), src.Spec.RefreshSchedule)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{Err: ingestErr, Retrigger: refreshOn}
}

// ingestionSpecHash computes a hash of only those source spec properties that impact ingestion.
func (r *SourceReconciler) ingestionSpecHash(spec *runtimev1.SourceSpec, srcConfig map[string]any) (string, error) {
	hash := md5.New()

	_, err := hash.Write([]byte(spec.SourceConnector))
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.SinkConnector))
	if err != nil {
		return "", err
	}

	st, err := structpb.NewStruct(srcConfig)
	if err != nil {
		return "", err
	}
	err = pbutil.WriteHash(structpb.NewStructValue(st), hash)
	if err != nil {
		return "", err
	}

	err = binary.Write(hash, binary.BigEndian, spec.TimeoutSeconds)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// stagingTableName returns a stable temporary table name for a destination table.
// By using a stable temporary table name, we can ensure proper garbage collection without managing additional state.
func (r *SourceReconciler) stagingTableName(table string) string {
	return "__rill_tmp_src_" + table
}

// setTriggerFalse sets the source's spec.Trigger to false.
// Unlike the State, the Spec may be edited concurrently with a Reconcile call, so we need to read and edit it under a lock.
func (r *SourceReconciler) setTriggerFalse(ctx context.Context, n *runtimev1.ResourceName) error {
	r.C.Lock(ctx)
	defer r.C.Unlock(ctx)

	self, err := r.C.Get(ctx, n, false)
	if err != nil {
		return err
	}

	source := self.GetSource()
	if source == nil {
		return fmt.Errorf("not a source")
	}

	source.Spec.Trigger = false
	return r.C.UpdateSpec(ctx, self.Meta.Name, self)
}

// ingestSource ingests the source into a table with tableName.
// It does NOT drop the table if ingestion fails after the table has been created.
// It will return an error if the sink connector is not an OLAP.
func (r *SourceReconciler) ingestSource(ctx context.Context, self *runtimev1.Resource, srcConfig, sinkConfig map[string]any) (outErr error) {
	src := self.GetSource().Spec

	// Get connections and transporter
	srcConn, release1, err := r.C.AcquireConn(ctx, src.SourceConnector)
	if err != nil {
		return err
	}
	defer release1()
	sinkConn, release2, err := r.C.AcquireConn(ctx, src.SinkConnector)
	if err != nil {
		return err
	}
	defer release2()
	t, ok := sinkConn.AsTransporter(srcConn, sinkConn)
	if !ok {
		t, ok = srcConn.AsTransporter(srcConn, sinkConn)
		if !ok {
			return fmt.Errorf("cannot transfer data between connectors %q and %q", src.SourceConnector, src.SinkConnector)
		}
	}

	// Set timeout on ctx
	timeout := _defaultIngestTimeout
	if src.TimeoutSeconds > 0 {
		timeout = time.Duration(src.TimeoutSeconds) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Get repo root
	repo, release, err := r.C.Runtime.Repo(ctx, r.C.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to access repo: %w", err)
	}
	repoRoot := repo.Root()
	release()

	// Execute the data transfer
	opts := &drivers.TransferOptions{
		AllowHostAccess: r.C.Runtime.AllowHostAccess(),
		RepoRoot:        repoRoot,
		AcquireConnector: func(name string) (drivers.Handle, func(), error) {
			return r.C.AcquireConn(ctx, name)
		},
		Progress: drivers.NoOpProgress{},
	}

	transferStart := time.Now()
	defer func() {
		transferLatency := time.Since(transferStart).Milliseconds()
		commonDims := []attribute.KeyValue{
			attribute.String("source", srcConn.Driver()),
			attribute.String("destination", sinkConn.Driver()),
			attribute.Bool("cancelled", errors.Is(outErr, context.Canceled)),
			attribute.Bool("failed", outErr != nil),
		}
		r.C.Activity.RecordMetric(ctx, "ingestion_ms", float64(transferLatency), commonDims...)

		// TODO: emit the number of bytes ingested (this might be extracted from a progress)
	}()

	err = t.Transfer(ctx, srcConfig, sinkConfig, opts)
	return err
}

func (r *SourceReconciler) driversSource(ctx context.Context, self *runtimev1.Resource, propsPB *structpb.Struct) (map[string]any, error) {
	tself := rillv1.TemplateResource{
		Meta:  self.Meta,
		Spec:  self.GetSource().Spec,
		State: self.GetSource().State,
	}

	return resolveTemplatedProps(ctx, r.C, tself, propsPB.AsMap())
}

func driversSink(tableName string) map[string]any {
	return map[string]any{"table": tableName}
}
