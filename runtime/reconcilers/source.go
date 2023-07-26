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
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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

func (r *SourceReconciler) stagingTableName(table string) string {
	return table + "_staging"
}

func (r *SourceReconciler) Reconcile(ctx context.Context, s *runtime.Signal) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, s.Name)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	src := self.GetSource()

	if self.Meta.Deleted {
		err1 := r.dropTableIfExists(ctx, src.State.Connector, src.State.Table)
		err2 := r.dropTableIfExists(ctx, src.State.Connector, r.stagingTableName(src.State.Table))
		err := errors.Join(err1, err2)
		return runtime.ReconcileResult{Err: err}
	}

	if self.Meta.RenamedFrom != nil {
		exists, err := r.tableExists(ctx, src.State.Connector, src.State.Table)
		if err == nil && exists {
			err = r.renameTable(ctx, src.State.Connector, src.State.Table, self.Meta.Name.Name)
		}
		if err != nil {
			return runtime.ReconcileResult{Err: fmt.Errorf("failed to rename table: %w", err)}
		}

		src.State.Table = self.Meta.Name.Name
		err = r.C.UpdateState(ctx, self.Meta.Name, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}

		return runtime.ReconcileResult{}
	}

	// TODO: Exit if refs have errors

	// Use a hash of ingestion-related fields from the spec to determine if we need to re-ingest
	hash, err := r.ingestionSpecHash(src.Spec)
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

	// Check if the table still exists (might have been dropped by the user)
	tableExists := false
	if src.State.Table != "" {
		tableExists, err = r.tableExists(ctx, src.State.Connector, src.State.Table)
		if err != nil {
			return runtime.ReconcileResult{Err: fmt.Errorf("could not check table: %w", err)}
		}
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

	// Do the actual ingestion
	table := self.Meta.Name.Name
	stagingTable := table
	connector := src.Spec.SinkConnector
	if src.Spec.StageChanges {
		stagingTable = r.stagingTableName(table)
	}
	ingestErr := r.ingestSource(ctx, src.Spec, stagingTable)
	if ingestErr != nil {
		ingestErr = fmt.Errorf("failed to ingest source: %w", ingestErr)
	}
	if ingestErr == nil && src.Spec.StageChanges {
		err = r.renameTable(ctx, connector, stagingTable, table)
		if err != nil {
			return runtime.ReconcileResult{Err: fmt.Errorf("failed to rename staging table: %w", err)}
		}
	}

	// How we handle ingestErr depends on StageChanges.
	// If StageChanges is true, we retain the existing table, but still return the error.
	// If StageChanges is false, we clear the table and return the error.

	// Update state
	update := false
	if ingestErr == nil {
		update = true
		src.State.Connector = connector
		src.State.Table = table
		src.State.SpecHash = hash
		src.State.RefreshedOn = timestamppb.Now()
	} else {
		if src.Spec.StageChanges {
			_ = r.dropTableIfExists(ctx, connector, stagingTable)
		} else {
			update = true
			_ = r.dropTableIfExists(ctx, connector, table)
			src.State.Connector = ""
			src.State.Table = ""
			src.State.SpecHash = ""
			src.State.RefreshedOn = nil
		}
	}
	if update {
		err = r.C.UpdateState(ctx, self.Meta.Name, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Reset spec.Trigger
	if src.Spec.Trigger {
		src.Spec.Trigger = false
		err = r.C.UpdateSpec(ctx, self.Meta.Name, self.Meta.Refs, self.Meta.Owner, self.Meta.FilePaths, self)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}
	}

	// Compute next refresh time
	refreshOn, err = nextRefreshTime(src.State.RefreshedOn.AsTime(), src.Spec.RefreshSchedule)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{Err: ingestErr, Retrigger: refreshOn}
}

// ingestionSpecHash computes a hash of only those source spec properties that impact ingestion.
func (r *SourceReconciler) ingestionSpecHash(spec *runtimev1.SourceSpec) (string, error) {
	hash := md5.New()

	_, err := hash.Write([]byte(spec.SourceConnector))
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(spec.SinkConnector))
	if err != nil {
		return "", err
	}

	err = pbutil.WriteHash(structpb.NewStructValue(spec.Properties), hash)
	if err != nil {
		return "", err
	}

	err = binary.Write(hash, binary.BigEndian, spec.TimeoutSeconds)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// tableExists returns true if the table exists in the database associated with connector
func (r *SourceReconciler) tableExists(ctx context.Context, connector string, table string) (bool, error) {
	if table == "" {
		return false, nil
	}

	panic("not implemented")
}

// renameTable renames the table from oldName to newName in the database associated with connector
func (r *SourceReconciler) renameTable(ctx context.Context, connector, oldName, newName string) error {
	panic("not implemented")
}

// dropTableIfExists drops the table from the database associated with connector
func (r *SourceReconciler) dropTableIfExists(ctx context.Context, connector string, table string) error {
	panic("not implemented")
}

// ingestSource ingests the source into a table with tableName
func (r *SourceReconciler) ingestSource(ctx context.Context, src *runtimev1.SourceSpec, tableName string) error {
	panic("not implemented")
}
