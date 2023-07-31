package sources

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
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
	return equal(cat1.GetSource().Properties.AsMap(), cat2.GetSource().Properties.AsMap())
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

func convertLower(in map[string]string) map[string]string {
	m := make(map[string]string, len(in))
	for key, value := range in {
		m[strings.ToLower(key)] = value
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

	var err error
	// TODO: this should go in the parser in the new reconcile
	if apiSource.Connector == "duckdb" {
		err = mergeFromParsedQuery(apiSource)
		if err != nil {
			return err
		}
	}

	logger = logger.With(zap.String("source", name))
	var srcConnector drivers.Connection

	if apiSource.Connector == "duckdb" {
		srcConnector = olap.(drivers.Connection)
	} else {
		var err error
		variables := convertLower(opts.InstanceEnv)
		srcConnector, err = drivers.Open(apiSource.Connector, connectorVariables(apiSource, variables, repo.Root()), logger)
		if err != nil {
			return fmt.Errorf("failed to open driver %w", err)
		}
		defer srcConnector.Close()
	}

	olapConnection := olap.(drivers.Connection)
	t, ok := olapConnection.AsTransporter(srcConnector, olapConnection)
	if !ok {
		t, ok = srcConnector.AsTransporter(srcConnector, olapConnection)
		if !ok {
			return fmt.Errorf("data transfer not possible from %q to %q", srcConnector.Driver(), olapConnection.Driver())
		}
	}

	src, err := source(apiSource.Connector, apiSource)
	if err != nil {
		return err
	}

	sink := sink(olapConnection.Driver(), name)

	timeout := _defaultIngestTimeout
	if apiSource.GetTimeoutSeconds() > 0 {
		timeout = time.Duration(apiSource.GetTimeoutSeconds()) * time.Second
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ingestionLimit := opts.IngestStorageLimitInBytes
	p := &progress{}
	limitExceeded := false
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctxWithTimeout.Done():
				return
			case <-ticker.C:
				olap, _ := olapConnection.AsOLAP()
				if size, ok := olap.EstimateSize(); ok && size > ingestionLimit {
					limitExceeded = true
					cancel()
				}
			}
		}
	}()
	err = t.Transfer(ctxWithTimeout, src, sink, drivers.NewTransferOpts(drivers.WithLimitInBytes(ingestionLimit)), p)
	if limitExceeded {
		return drivers.ErrIngestionLimitExceeded
	}
	return err
}

func mergeFromParsedQuery(apiSource *runtimev1.Source) error {
	props := apiSource.Properties.AsMap()
	query, ok := props["sql"]
	if !ok {
		return nil
	}
	queryStr, ok := query.(string)
	if !ok {
		return errors.New("query should be a string")
	}

	// raw sql query
	ast, err := duckdbsql.Parse(queryStr)
	if err != nil {
		return err
	}
	refs := ast.GetTableRefs()
	if len(refs) != 1 {
		return errors.New("sql source should have exactly one table reference")
	}
	ref := refs[0]

	if len(ref.Paths) == 0 {
		return errors.New("only read_* functions with a single path is supported")
	}
	if len(ref.Paths) > 1 {
		return errors.New("invalid source, only a single path for source is supported")
	}

	p, c, ok := parseEmbeddedSourceConnector(ref.Paths[0])
	if !ok {
		return errors.New("unknown source")
	}
	if c == "local_file" {
		return nil
	}

	apiSource.Connector = c
	props["path"] = p
	props["sql"] = queryStr

	pbProps, err := structpb.NewStruct(props)
	if err != nil {
		return err
	}
	apiSource.Properties = pbProps
	return nil
}

type progress struct {
	catalogObj drivers.CatalogEntry
	unit       drivers.ProgressUnit
}

func (p *progress) Target(val int64, unit drivers.ProgressUnit) {
	p.unit = unit
}

func (p *progress) Observe(val int64, unit drivers.ProgressUnit) {
	if unit == drivers.ProgressUnitByte {
		p.catalogObj.BytesIngested += val
	}
}

func source(connector string, src *runtimev1.Source) (drivers.Source, error) {
	props := src.Properties.AsMap()
	switch connector {
	case "s3":
		return &drivers.BucketSource{
			ExtractPolicy: src.Policy,
			Properties:    props,
		}, nil
	case "gcs":
		return &drivers.BucketSource{
			ExtractPolicy: src.Policy,
			Properties:    props,
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
		return nil, fmt.Errorf("connector %v not supported", connector)
	}
}

func sink(connector, tableName string) drivers.Sink {
	switch connector {
	case "duckdb":
		return &drivers.DatabaseSink{
			Table: tableName,
		}
	default:
		return nil
	}
}

func connectorVariables(src *runtimev1.Source, env map[string]string, repoRoot string) map[string]any {
	connector := src.Connector
	vars := map[string]any{
		"allow_host_access": strings.EqualFold(env["allow_host_access"], "true"),
	}
	switch connector {
	case "s3":
		vars["aws_access_key_id"] = env["aws_access_key_id"]
		vars["aws_secret_access_key"] = env["aws_secret_access_key"]
		vars["aws_session_token"] = env["aws_session_token"]
	case "gcs":
		vars["google_application_credentials"] = env["google_application_credentials"]
	case "motherduck":
		vars["token"] = env["token"]
		vars["dsn"] = ""
	case "local_file":
		vars["dsn"] = repoRoot
	case "bigquery":
		vars["google_application_credentials"] = env["google_application_credentials"]
	}
	return vars
}

func equal(s, o map[string]any) bool {
	return reflect.DeepEqual(s, o)
}
