package sources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
)

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
) error {
	return ingestSource(ctx, olap, repo, opts, catalogObj, "")
}

func (m *sourceMigrator) Update(ctx context.Context,
	olap drivers.OLAPStore,
	repo drivers.RepoStore,
	opts migrator.Options,
	oldCatalogObj, newCatalogObj *drivers.CatalogEntry,
) error {
	apiSource := newCatalogObj.GetSource()

	tempName := fmt.Sprintf("__rill_temp_%s", apiSource.Name)

	err := ingestSource(ctx, olap, repo, opts, newCatalogObj, tempName)
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

func ingestSource(
	ctx context.Context,
	olap drivers.OLAPStore,
	repo drivers.RepoStore,
	opts migrator.Options,
	catalogObj *drivers.CatalogEntry,
	name string,
) error {
	apiSource := catalogObj.GetSource()

	if name == "" {
		name = apiSource.Name
	}

	source := &connectors.Source{
		Name:          name,
		Connector:     apiSource.Connector,
		Properties:    apiSource.Properties.AsMap(),
		ExtractPolicy: apiSource.GetPolicy(),
		Timeout:       apiSource.GetTimeoutSeconds(),
	}

	variables := convertUpper(opts.InstanceEnv)
	env := &connectors.Env{
		RepoDriver:           repo.Driver(),
		RepoRoot:             repo.Root(),
		Variables:            variables,
		AllowHostCredentials: strings.EqualFold(variables["ALLOW_HOST_CREDENTIALS"], "true"),
		StorageLimitInBytes:  opts.IngestStorageLimitInBytes,
		DisableAbsolutePath:  strings.EqualFold(variables["DISABLE_ABSOLUTE_PATH"], "true"),
	}

	ingestionSummary, err := olap.Ingest(ctx, env, source)
	if err != nil {
		return err
	}

	catalogObj.BytesIngested = ingestionSummary.BytesIngested
	return nil
}
