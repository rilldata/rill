package reconcilers

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
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

func (r *SourceReconciler) Reconcile(ctx context.Context, s *runtime.Signal) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, s.Name)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	src := self.GetSource()

	// Handle deletion
	if self.Meta.Deleted {
		err1 := r.dropTableIfExists(ctx, src.State.Connector, src.State.Table)
		err2 := r.dropTableIfExists(ctx, src.State.Connector, r.stagingTableName(src.State.Table))
		err := errors.Join(err1, err2)
		return runtime.ReconcileResult{Err: err}
	}

	// Handle renames
	tableName := self.Meta.Name.Name
	if self.Meta.RenamedFrom != nil {
		exists, err := r.tableExists(ctx, src.State.Connector, src.State.Table)
		if err != nil {
			return runtime.ReconcileResult{Err: err}
		}

		if exists {
			err = r.renameTable(ctx, src.State.Connector, src.State.Table, tableName)
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

	// If the SinkConnector was changed, drop data in the old connector
	if src.State.Table != "" && src.State.Connector != src.Spec.SinkConnector {
		_ = r.dropTableIfExists(ctx, src.State.Connector, src.State.Table)
		_ = r.dropTableIfExists(ctx, src.State.Connector, r.stagingTableName(src.State.Table))
	}

	// Do the actual ingestion
	stagingTableName := tableName
	connector := src.Spec.SinkConnector
	if src.Spec.StageChanges {
		stagingTableName = r.stagingTableName(tableName)
	}
	ingestErr := r.ingestSource(ctx, src.Spec, stagingTableName)
	if ingestErr != nil {
		ingestErr = fmt.Errorf("failed to ingest source: %w", ingestErr)
	}
	if ingestErr == nil && src.Spec.StageChanges {
		err = r.renameTable(ctx, connector, stagingTableName, tableName)
		if err != nil {
			return runtime.ReconcileResult{Err: fmt.Errorf("failed to rename staging table: %w", err)}
		}
	}

	// How we handle ingestErr depends on StageChanges.
	// If StageChanges is true, we retain the existing table, but still return the error.
	// If StageChanges is false, we clear the existing table and return the error.

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
		_ = r.dropTableIfExists(ctx, connector, stagingTableName)
	} else {
		// Failed ingestion to main table
		update = true
		_ = r.dropTableIfExists(ctx, connector, tableName)
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

	// Reset spec.Trigger
	if src.Spec.Trigger {
		src.Spec.Trigger = false
		err = r.C.UpdateSpec(ctx, self.Meta.Name, self.Meta.Refs, self.Meta.Owner, self.Meta.FilePaths, self)
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

// stagingTableName returns a stable temporary table name for a destination table.
// By using a stable temporary table name, we can ensure proper garbage collection without managing additional state.
func (r *SourceReconciler) stagingTableName(table string) string {
	return "__rill_tmp_src_" + table
}

// tableExists returns true if the table exists in the database associated with connector.
// It will return an error if connector is not an OLAP.
func (r *SourceReconciler) tableExists(ctx context.Context, connector, table string) (bool, error) {
	if table == "" {
		return false, nil
	}

	olap, release, err := r.C.AcquireOLAP(ctx, connector)
	if err != nil {
		return false, err
	}
	defer release()

	_, err = olap.InformationSchema().Lookup(ctx, table)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// renameTable renames the table from oldName to newName in the database associated with connector.
// If newName already exists, it is overwritten.
// It will return an error if connector is not an OLAP.
func (r *SourceReconciler) renameTable(ctx context.Context, connector, oldName, newName string) error {
	if oldName == "" || newName == "" {
		return fmt.Errorf("cannot rename empty table name: oldName=%q, newName=%q", oldName, newName)
	}

	if oldName == newName {
		return nil
	}

	olap, release, err := r.C.AcquireOLAP(ctx, connector)
	if err != nil {
		return err
	}
	defer release()

	return olap.WithConnection(ctx, 100, func(ctx context.Context, ensuredCtx context.Context) error {
		// TODO: Use a transaction?

		// DuckDB does not support renaming a table to the same name with different casing.
		// Workaround by renaming to a temporary name first.
		if strings.EqualFold(oldName, newName) {
			n := r.stagingTableName(newName)
			err = olap.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("DROP TABLE IF EXISTS %s", safeSQLName(n))})
			if err != nil {
				return err
			}

			err := olap.Exec(ctx, &drivers.Statement{
				Query:    fmt.Sprintf("ALTER TABLE %s RENAME TO %s", safeSQLName(oldName), safeSQLName(n)),
				Priority: 100,
			})
			if err != nil {
				return err
			}
			oldName = n
		}

		err = olap.Exec(ctx, &drivers.Statement{Query: fmt.Sprintf("DROP TABLE IF EXISTS %s", safeSQLName(newName))})
		if err != nil {
			return err
		}

		return olap.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("ALTER TABLE %s RENAME TO %s", safeSQLName(oldName), safeSQLName(newName)),
			Priority: 100,
		})
	})
}

// dropTableIfExists drops the table from the database associated with connector.
// It will return an error if connector is not an OLAP.
func (r *SourceReconciler) dropTableIfExists(ctx context.Context, connector, table string) error {
	if table == "" {
		return nil
	}

	olap, release, err := r.C.AcquireOLAP(ctx, connector)
	if err != nil {
		return err
	}
	defer release()

	return olap.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("DROP TABLE IF EXISTS %s", safeSQLName(table)),
		Priority: 100,
	})
}

// ingestSource ingests the source into a table with tableName.
// It does NOT drop the table if ingestion fails after the table has been created.
// It will return an error if the sink connector is not an OLAP.
func (r *SourceReconciler) ingestSource(ctx context.Context, src *runtimev1.SourceSpec, tableName string) error {
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

	// Get source and sink configs
	srcConfig, err := driversSource(srcConn, src.Properties)
	if err != nil {
		return err
	}
	sinkConfig, err := driversSink(sinkConn, tableName)
	if err != nil {
		return err
	}

	// Set timeout on ctx
	timeout := _defaultIngestTimeout
	if src.TimeoutSeconds > 0 {
		timeout = time.Duration(src.TimeoutSeconds) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Enforce storage limits
	// TODO: This code is pretty ugly. We should push storage limit tracking into the underlying driver and transporter.
	var ingestionLimit *int64
	var limitExceeded bool
	if olap, ok := sinkConn.AsOLAP(); ok {
		// Get storage limit
		inst, err := r.C.Runtime.FindInstance(ctx, r.C.InstanceID)
		if err != nil {
			return err
		}
		storageLimit := inst.IngestionLimitBytes

		// Enforce storage limit if it's set
		if storageLimit > 0 {
			// Get ingestion limit (storage limit minus current size)
			bytes, ok := olap.EstimateSize()
			if ok {
				n := storageLimit - bytes
				if n <= 0 {
					return drivers.ErrIngestionLimitExceeded
				}
				ingestionLimit = &n

				// Start background goroutine to check size is not exceeded during ingestion
				go func() {
					ticker := time.NewTicker(5 * time.Second)
					defer ticker.Stop()
					for {
						select {
						case <-ctx.Done():
							return
						case <-ticker.C:
							if size, ok := olap.EstimateSize(); ok && size > storageLimit {
								limitExceeded = true
								cancel()
							}
						}
					}
				}()
			}
		}
	}

	// Execute the data transfer
	opts := drivers.NewTransferOpts()
	if ingestionLimit != nil {
		opts.LimitInBytes = *ingestionLimit
	}
	err = t.Transfer(ctx, srcConfig, sinkConfig, opts, drivers.NoOpProgress{})
	if limitExceeded {
		return drivers.ErrIngestionLimitExceeded
	}
	return err
}

func driversSource(conn drivers.Connection, propsPB *structpb.Struct) (drivers.Source, error) {
	props := propsPB.AsMap()
	switch conn.Driver() {
	case "s3":
		return &drivers.BucketSource{
			// ExtractPolicy: src.Policy, // TODO: Add
			Properties: props,
		}, nil
	case "gcs":
		return &drivers.BucketSource{
			// ExtractPolicy: src.Policy, // TODO: Add
			Properties: props,
		}, nil
	case "https":
		return &drivers.FileSource{
			Properties: props,
		}, nil
	case "local_file":
		return &drivers.FileSource{
			Properties: props,
		}, nil
	case "motherduck":
		query, ok := props["sql"].(string)
		if !ok {
			return nil, fmt.Errorf("property \"sql\" is mandatory for connector \"motherduck\"")
		}
		var db string
		if val, ok := props["db"].(string); ok {
			db = val
		}

		return &drivers.DatabaseSource{
			SQL:      query,
			Database: db,
		}, nil
	case "duckdb":
		query, ok := props["sql"].(string)
		if !ok {
			return nil, fmt.Errorf("property \"sql\" is mandatory for connector \"duckdb\"")
		}
		return &drivers.DatabaseSource{
			SQL: query,
		}, nil
	case "bigquery":
		query, ok := props["sql"].(string)
		if !ok {
			return nil, fmt.Errorf("property \"sql\" is mandatory for connector \"bigquery\"")
		}
		return &drivers.DatabaseSource{
			SQL:   query,
			Props: props,
		}, nil
	default:
		return nil, fmt.Errorf("source connector %q not supported", conn.Driver())
	}
}

func driversSink(conn drivers.Connection, tableName string) (drivers.Sink, error) {
	switch conn.Driver() {
	case "duckdb":
		return &drivers.DatabaseSink{
			Table: tableName,
		}, nil
	default:
		return nil, fmt.Errorf("sink connector %q not supported", conn.Driver())
	}
}

/*
Features in old migrator not yet addressed:
- parsing ExtractPolicy
- rewriting source if `sql` field is set (mergeFromParsedQuery)
- opening srcConnector with correct variables (connectorVariables)
- adding source name to logger, and propagating the controller logger to the connection/transporter
*/
