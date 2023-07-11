package sources

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
	"go.uber.org/zap"
)

const _defaultIngestTimeout = 60 * time.Minute

func init() {
	migrator.Register(drivers.ObjectTypeSource, &sourceMigrator{})
}

type sourceMigrator struct{}

func (m *sourceMigrator) Create(
	ctx context.Context,
	olap drivers.OLAPStore,
	repo drivers.RepoStore,
	opts migrator.Options,
	catalogObj *drivers.CatalogEntry,
	logger *zap.Logger,
) error {
	return ingestSource(ctx, olap, repo, opts, catalogObj, "", logger)
}

func (m *sourceMigrator) Update(ctx context.Context,
	olap drivers.OLAPStore,
	repo drivers.RepoStore,
	opts migrator.Options,
	oldCatalogObj, newCatalogObj *drivers.CatalogEntry,
	logger *zap.Logger,
) error {
	apiSource := newCatalogObj.GetSource()

	tempName := fmt.Sprintf("__rill_temp_%s", apiSource.Name)

	err := ingestSource(ctx, olap, repo, opts, newCatalogObj, tempName, logger)
	if err != nil {
		// cleanup of temp table. can exist and still error out in incremental ingestion
		_ = olap.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP TABLE IF EXISTS %s", tempName),
			Priority: 100,
		})
		// return the original error. error for dropping is less important for the user
		return err
	}

	tempNameOrig := fmt.Sprintf("__rill_temp_orig_%s", apiSource.Name)
	// drop the temp for original if exists
	err = olap.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("DROP TABLE IF EXISTS %s", tempNameOrig),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	// rename the original to temp original table
	err = olap.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("ALTER TABLE %s RENAME TO %s", apiSource.Name, tempNameOrig),
		Priority: 100,
	})
	if err != nil {
		return err
	}

	// finally rename the new temp table to actual table
	err = olap.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("ALTER TABLE %s RENAME TO %s", tempName, apiSource.Name),
		Priority: 100,
	})
	if err != nil {
		// revert the original table
		_ = olap.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("ALTER TABLE %s RENAME TO %s", apiSource.Name, tempNameOrig),
			Priority: 100,
		})

		// cleanup of temp table
		_ = olap.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("DROP TABLE IF EXISTS %s", tempName),
			Priority: 100,
		})
		// original error is more important that the error from drop of temp table
		return err
	}

	// cleanup the backup of original
	err = olap.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("DROP TABLE %s", tempNameOrig),
		Priority: 100,
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *sourceMigrator) Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalogObj *drivers.CatalogEntry) error {
	if strings.EqualFold(from, catalogObj.Name) {
		tempName := fmt.Sprintf("__rill_temp_%s", from)
		err := olap.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("ALTER TABLE %s RENAME TO %s", from, tempName),
			Priority: 100,
		})
		if err != nil {
			return err
		}
		from = tempName
	}

	return olap.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("ALTER TABLE %s RENAME TO %s", from, catalogObj.Name),
		Priority: 100,
	})
}

func (m *sourceMigrator) Delete(ctx context.Context, olap drivers.OLAPStore, catalogObj *drivers.CatalogEntry) error {
	return olap.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("DROP TABLE IF EXISTS %s", catalogObj.Name),
		Priority: 100,
	})
}

func (m *sourceMigrator) GetDependencies(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) ([]string, []*drivers.CatalogEntry) {
	return []string{}, nil
}

func (m *sourceMigrator) Validate(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) []*runtimev1.ReconcileError {
	// TODO - Details needs to be added here
	return nil
}

func (m *sourceMigrator) IsEqual(ctx context.Context, cat1, cat2 *drivers.CatalogEntry) bool {
	if cat1.GetSource().Connector != cat2.GetSource().Connector {
		return false
	}
	if !comparePolicy(cat1.GetSource().GetPolicy(), cat2.GetSource().GetPolicy()) {
		return false
	}
	s1 := &connectors.Source{
		Properties: cat1.GetSource().Properties.AsMap(),
	}
	s2 := &connectors.Source{
		Properties: cat2.GetSource().Properties.AsMap(),
	}
	return s1.PropertiesEquals(s2)
}

func comparePolicy(p1, p2 *runtimev1.Source_ExtractPolicy) bool {
	if (p1 != nil) == (p2 != nil) {
		if p1 != nil {
			// both non nil
			return p1.FilesStrategy == p2.FilesStrategy &&
				p1.FilesLimit == p2.FilesLimit &&
				p1.RowsStrategy == p2.RowsStrategy &&
				p1.RowsLimitBytes == p2.RowsLimitBytes
		}
		// both nil
		return true
	}
	return false
}

func (m *sourceMigrator) ExistsInOlap(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) (bool, error) {
	_, err := olap.InformationSchema().Lookup(ctx, catalog.Name)
	if errors.Is(err, drivers.ErrNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func convertUpper(in map[string]string) map[string]string {
	m := make(map[string]string, len(in))
	for key, value := range in {
		m[strings.ToUpper(key)] = value
	}
	return m
}

func ingestSource(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, opts migrator.Options,
	catalogObj *drivers.CatalogEntry, name string, logger *zap.Logger,
) error {
	apiSource := catalogObj.GetSource()

	if name == "" {
		name = apiSource.Name
	}

	// variables := convertUpper(opts.InstanceEnv)
	// env := &connectors.Env{
	// 	RepoDriver:          repo.Driver(),
	// 	RepoRoot:            repo.Root(),
	// 	Variables:           variables,
	// 	AllowHostAccess:     strings.EqualFold(variables["ALLOW_HOST_ACCESS"], "true"),
	// 	StorageLimitInBytes: opts.IngestStorageLimitInBytes,
	// }

	logger = logger.With(zap.String("source", name))
	variables := convertUpper(opts.InstanceEnv)
	connector, err := drivers.Open(apiSource.Connector, connectorVariables(apiSource.Connector, variables), logger)
	if err != nil {
		return fmt.Errorf("failed to open driver %w", err)
	}

	olapConnection := olap.(drivers.Connection)
	t, ok := olapConnection.AsTransporter(connector, olapConnection)
	if !ok {
		t, ok = connector.AsTransporter(connector, olapConnection)
		if !ok {
			return fmt.Errorf("data transfer not possible from %q to %q", connector.Driver(), olapConnection.Driver())
		}
	}

	src := source(apiSource.Connector, apiSource)
	sink := sink(olapConnection.Driver(), name)

	timeout := _defaultIngestTimeout
	if apiSource.GetTimeoutSeconds() > 0 {
		timeout = time.Duration(apiSource.GetTimeoutSeconds()) * time.Second
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	p := &progress{}
	err = t.Transfer(ctxWithTimeout, src, sink, drivers.NewTransferOpts(drivers.WithLimitInBytes(opts.IngestStorageLimitInBytes)), p)
	if err != nil {
		return err
	}

	catalogObj.BytesIngested = p.size
	return nil
}

type progress struct {
	size int64
	unit drivers.ProgressUnit
}

func (p *progress) Target(val int64, unit drivers.ProgressUnit) {
	p.unit = unit
}

func (p *progress) Observe(val int64, unit drivers.ProgressUnit) {
	p.size += val
}

func source(connector string, src *runtimev1.Source) drivers.Source {
	props := src.Properties.AsMap()
	switch connector {
	case "s3":
		return &drivers.BucketSource{
			Paths: []string{
				props["uri"].(string),
			},
			ExtractPolicy: src.Policy,
			Properties:    props,
		}
	case "gcs":
		return &drivers.BucketSource{
			Paths: []string{
				props["uri"].(string),
			},
			ExtractPolicy: src.Policy,
			Properties:    props,
		}
	case "http":
		return &drivers.FilesSource{
			Properties: props,
		}
	case "local_file":
		return &drivers.BucketSource{
			Properties: props,
		}
	default:
		return nil
	}
}

func sink(connector string, tableName string) drivers.Sink {
	switch connector {
	case "duckdb":
		return &drivers.DatabaseSink{
			Table: tableName,
		}
	default:
		return nil
	}
}

func connectorVariables(connector string, env map[string]string) map[string]any {
	vars := map[string]any{
		"ALLOW_HOST_ACCESS": env["ALLOW_HOST_ACCESS"],
	}
	switch connector {
	case "s3":
		vars["AWS_ACCESS_KEY_ID"] = env["AWS_ACCESS_KEY_ID"]
		vars["AWS_SECRET_ACCESS_KEY"] = env["AWS_SECRET_ACCESS_KEY"]
		vars["AWS_SESSION_TOKEN"] = env["AWS_SESSION_TOKEN"]
	case "gcs":
		vars["GOOGLE_APPLICATION_CREDENTIALS"] = env["GOOGLE_APPLICATION_CREDENTIALS"]
		vars["ALLOW_HOST_ACCESS"] = env["ALLOW_HOST_ACCESS"]
	case "motherduck":
		vars["TOKEN"] = env["TOKEN"]
		vars["dsn"] = ""
	}
	return vars
}
