package catalog_test

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/file"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/rilldata/rill/runtime/services/catalog"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/sql"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/yaml"
	"github.com/rilldata/rill/runtime/services/catalog/migrator/metricsviews"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/models"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
	"github.com/rilldata/rill/runtime/services/catalog/testutils"
	"github.com/stretchr/testify/require"
)

const TestDataPath = "../../../web-local/test/data"

var AdBidsCsvPath = filepath.Join(TestDataPath, "AdBids.csv")
var AdImpressionsCsvPath = filepath.Join(TestDataPath, "AdImpressions.tsv")
var BrokenCsvPath = filepath.Join(TestDataPath, "BrokenCSV.csv")

const AdBidsRepoPath = "/sources/AdBids.yaml"
const AdImpressionsRepoPath = "/sources/AdImpressions.yaml"
const AdBidsNewRepoPath = "/sources/AdBidsNew.yaml"
const AdBidsModelRepoPath = "/models/AdBids_model.sql"
const AdBidsSourceModelRepoPath = "/models/AdBids_source_model.sql"
const AdBidsDashboardRepoPath = "/dashboards/AdBids_dashboard.yaml"

var AdBidsAffectedPaths = []string{AdBidsRepoPath, AdBidsModelRepoPath, AdBidsDashboardRepoPath}
var AdBidsNewAffectedPaths = []string{AdBidsNewRepoPath, AdBidsModelRepoPath, AdBidsDashboardRepoPath}
var AdBidsDashboardAffectedPaths = []string{AdBidsModelRepoPath, AdBidsDashboardRepoPath}

func TestReconcile(t *testing.T) {
	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			// same name different content
			testutils.CreateSource(t, s, "AdBids", AdImpressionsCsvPath, AdBidsRepoPath)
			result, err := s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 2, 0, 1, 0, AdBidsAffectedPaths)
			require.Equal(t, metricsviews.SourceNotFound, result.Errors[1].Message)
			testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)
			testutils.AssertTableAbsence(t, s, "AdBids_model")

			// revert to stable state
			testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			// TODO: should the model/dashboard be counted as updated or added
			testutils.AssertMigration(t, result, 0, 2, 1, 0, AdBidsAffectedPaths)
			testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

			// update with same content
			testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 0, 0, 0, []string{})

			// delete from olap
			res, err := s.Olap.Execute(context.Background(), &drivers.Statement{
				Query: "drop table AdBids",
			})
			require.NoError(t, err)
			require.NoError(t, res.Close())
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 1, 2, 0, AdBidsAffectedPaths)

			// delete file
			err = os.Remove(path.Join(dir, AdBidsRepoPath))
			require.NoError(t, err)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 2, 0, 0, 1, AdBidsAffectedPaths)
			testutils.AssertTableAbsence(t, s, "AdBids")
			testutils.AssertTableAbsence(t, s, "AdBids_model")

			// add back source
			testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 3, 0, 0, AdBidsAffectedPaths)
			testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)
		})
	}
}

func TestReconcileRenames(t *testing.T) {
	if testing.Short() {
		t.Skip("renames: skipping test in short mode")
	}
	AdBidsCapsRepoPath := "/sources/ADBIDS.yaml"

	configs := []struct {
		title               string
		config              catalog.ReconcileConfig
		configForCaseChange catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}, catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsRepoPath, AdBidsNewRepoPath},
		}, catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsRepoPath, AdBidsNewRepoPath, AdBidsCapsRepoPath},
		}},
		{"ReconcileSelectedReverseOrder", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsNewRepoPath, AdBidsRepoPath},
		}, catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsCapsRepoPath, AdBidsNewRepoPath, AdBidsRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			// write to a new file (should rename)
			testutils.RenameFile(t, dir, AdBidsRepoPath, AdBidsNewRepoPath)
			result, err := s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 2, 0, 1, 0, AdBidsNewAffectedPaths)
			testutils.AssertTableAbsence(t, s, "AdBids")
			testutils.AssertTable(t, s, "AdBidsNew", AdBidsNewRepoPath)
			testutils.AssertTableAbsence(t, s, "AdBids_model")

			// write to the previous file (should rename back to original)
			testutils.RenameFile(t, dir, AdBidsNewRepoPath, AdBidsRepoPath)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 2, 1, 0, AdBidsAffectedPaths)
			testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)
			testutils.AssertTableAbsence(t, s, "AdBidsNew")
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

			AdBidsCapsAffectedPaths := []string{AdBidsCapsRepoPath, AdBidsModelRepoPath, AdBidsDashboardRepoPath}
			// write to a new file with same name and different case
			testutils.RenameFile(t, dir, AdBidsRepoPath, AdBidsCapsRepoPath)
			result, err = s.Reconcile(context.Background(), tt.configForCaseChange)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 0, 3, 0, AdBidsCapsAffectedPaths)
			testutils.AssertTable(t, s, "ADBIDS", AdBidsCapsRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

			// update with same content
			testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
			result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{
				ChangedPaths: []string{AdBidsCapsRepoPath},
				ForcedPaths:  []string{AdBidsCapsRepoPath},
			})
			require.NoError(t, err)
			// ForcedPaths updates all dependant items
			testutils.AssertMigration(t, result, 0, 0, 3, 0, AdBidsCapsAffectedPaths)
		})
	}
}

func TestRefreshSource(t *testing.T) {
	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{
			ForcedPaths: []string{AdBidsRepoPath},
		}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ForcedPaths:  []string{AdBidsRepoPath},
			ChangedPaths: []string{AdBidsRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			testutils.CopyFileToData(t, dir, AdBidsCsvPath, "AdBids.csv")
			AdBidsDataPath := "data/AdBids.csv"

			// update with same content
			err := artifacts.Write(context.Background(), s.Repo, s.InstID, &drivers.CatalogEntry{
				Name: "AdBids",
				Type: drivers.ObjectTypeSource,
				Path: AdBidsRepoPath,
				Object: &runtimev1.Source{
					Name:      "AdBids",
					Connector: "local_file",
					Properties: testutils.ToProtoStruct(map[string]any{
						"path": AdBidsDataPath,
					}),
				},
			})
			require.NoError(t, err)
			result, err := s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			// ForcedPaths updates all dependant items
			testutils.AssertMigration(t, result, 0, 0, 3, 0, AdBidsAffectedPaths)

			// update the uploaded file directly
			time.Sleep(10 * time.Millisecond)
			err = os.Chtimes(path.Join(dir, AdBidsDataPath), time.Now(), time.Now())
			require.NoError(t, err)
			result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{
				ChangedPaths: tt.config.ChangedPaths,
			})
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 0, 3, 0, AdBidsAffectedPaths)

			// refresh with invalid data
			time.Sleep(10 * time.Millisecond)
			testutils.CopyFileToData(t, dir, BrokenCsvPath, "AdBids.csv")
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 2, 0, 1, 0, AdBidsAffectedPaths)

			// refresh again with valid data
			time.Sleep(10 * time.Millisecond)
			testutils.CopyFileToData(t, dir, AdBidsCsvPath, "AdBids.csv")
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 2, 1, 0, AdBidsAffectedPaths)
		})
	}
}

func TestInterdependentModel(t *testing.T) {
	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsRepoPath},
		}},
	}

	AdBidsSourceAffectedPaths := []string{AdBidsSourceModelRepoPath, AdBidsModelRepoPath, AdBidsDashboardRepoPath}
	AdBidsAllAffectedPaths := append([]string{AdBidsRepoPath}, AdBidsSourceAffectedPaths...)

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, _ := initBasicService(t)

			testutils.CreateModel(t, s, "AdBids_source_model",
				"select id, timestamp, publisher, domain, bid_price from AdBids", AdBidsSourceModelRepoPath)
			testutils.CreateModel(t, s, "AdBids_model",
				"select id, timestamp, publisher, domain, bid_price from AdBids_source_model", AdBidsModelRepoPath)
			result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{})
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 1, 2, 0, AdBidsSourceAffectedPaths)
			testutils.AssertTable(t, s, "AdBids_source_model", AdBidsSourceModelRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

			// trigger error in source
			testutils.CreateSource(t, s, "AdBids", AdImpressionsCsvPath, AdBidsRepoPath)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 3, 0, 1, 0, AdBidsAllAffectedPaths)
			require.Equal(t, metricsviews.SourceNotFound, result.Errors[2].Message)
			testutils.AssertTableAbsence(t, s, "AdBids_source_model")
			testutils.AssertTableAbsence(t, s, "AdBids_model")

			// reset the source
			testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 3, 1, 0, AdBidsAllAffectedPaths)
			testutils.AssertTable(t, s, "AdBids_source_model", AdBidsSourceModelRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)
		})
	}
}

func TestModelRename(t *testing.T) {
	var AdBidsRenameModelRepoPath = "/models/AdBidsRename.sql"
	var AdBidsRenameNewModelRepoPath = "/models/AdBidsRenameNew.sql"

	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsRenameModelRepoPath, AdBidsRenameNewModelRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			testutils.CreateModel(t, s, "AdBidsRename", "select * from AdBids", AdBidsRenameModelRepoPath)
			result, err := s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 1, 0, 0, []string{AdBidsRenameModelRepoPath})

			for i := 0; i < 5; i++ {
				testutils.RenameFile(t, dir, AdBidsRenameModelRepoPath, AdBidsRenameNewModelRepoPath)
				result, err = s.Reconcile(context.Background(), tt.config)
				require.NoError(t, err)
				testutils.AssertMigration(t, result, 0, 0, 1, 0, []string{AdBidsRenameNewModelRepoPath})

				testutils.RenameFile(t, dir, AdBidsRenameNewModelRepoPath, AdBidsRenameModelRepoPath)
				result, err = s.Reconcile(context.Background(), tt.config)
				require.NoError(t, err)
				testutils.AssertMigration(t, result, 0, 0, 1, 0, []string{AdBidsRenameModelRepoPath})
			}
		})
	}
}

func TestModelRenameToSource(t *testing.T) {
	var AdBidsModelAsSource = "/models/AdBids.sql"

	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsModelRepoPath, AdBidsModelAsSource},
		}},
		{"ReconcileSelectedReversed", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsModelAsSource, AdBidsModelRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			testutils.RenameFile(t, dir, AdBidsModelRepoPath, AdBidsModelAsSource)
			result, err := s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 2, 0, 0, 1,
				[]string{AdBidsModelRepoPath, AdBidsModelAsSource, AdBidsDashboardRepoPath})
			require.Equal(t, "item with same name exists", result.Errors[0].Message)
			testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)
			testutils.AssertTableAbsence(t, s, "AdBids_model")

			// reset state
			testutils.RenameFile(t, dir, AdBidsModelAsSource, AdBidsModelRepoPath)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			errCount := 0
			changedPaths := []string{AdBidsModelRepoPath, AdBidsDashboardRepoPath}
			if len(tt.config.ChangedPaths) > 0 {
				errCount = 1
				changedPaths = append(changedPaths, AdBidsModelAsSource)
			}
			// TODO: fix the issue of AdBidsModelAsSource being marked as error here
			testutils.AssertMigration(t, result, errCount, 2, 0, 0, changedPaths)
			testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)
			testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)
		})
	}
}

func TestModelVariations(t *testing.T) {
	s, _ := initBasicService(t)

	// same query with spaces
	testutils.CreateModel(t, s, "AdBids_model",
		`
-- this is a comment
select id,   timestamp,publisher, domain,
bid_price from AdBids;
`, AdBidsModelRepoPath)
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	// no change
	testutils.AssertMigration(t, result, 0, 0, 0, 0, []string{})

	// update to invalid model
	testutils.CreateModel(t, s, "AdBids_model",
		"select id, timestamp, publisher, domain, bid_price AdBids", AdBidsModelRepoPath)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 2, 0, 0, 0, AdBidsDashboardAffectedPaths)
	testutils.AssertTableAbsence(t, s, "AdBids_model")

	// new invalid model
	testutils.CreateModel(t, s, "AdBids_source_model",
		"select id, timestamp, publisher, domain, bid_price AdBids", AdBidsSourceModelRepoPath)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 1, 0, 0, 0, []string{AdBidsSourceModelRepoPath})
	testutils.AssertTableAbsence(t, s, "AdBids_source_model")
}

func TestModelWithMissingSource(t *testing.T) {
	s, _ := initBasicService(t)

	testutils.CreateModel(t, s, "AdBids_model", "select * from AdImpressions", AdBidsSourceModelRepoPath)
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 1, 0, 0, 0, []string{AdBidsSourceModelRepoPath})

	// update with a CTE with missing alias but valid and existing source
	testutils.CreateModel(t, s, "AdBids_model",
		"with CTEAlias as (select * from AdBids) select * from CTEAlias", AdBidsSourceModelRepoPath)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 1, 0, 0, []string{AdBidsSourceModelRepoPath})

	// update source with same content
	testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{
		// force update to test DAG
		ForcedPaths: []string{AdBidsRepoPath},
	})
	require.NoError(t, err)
	// changes propagate to model
	testutils.AssertMigration(t, result, 0, 0, 4, 0,
		append([]string{AdBidsSourceModelRepoPath}, AdBidsAffectedPaths...))
}

func TestReconcileMetricsView(t *testing.T) {
	s, _ := initBasicService(t)

	testutils.CreateModel(t, s, "AdBids_model", "select id, publisher, domain, bid_price from AdBids", AdBidsModelRepoPath)
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 1, 0, 1, 0, AdBidsDashboardAffectedPaths)
	// dropping the timestamp column gives a different error
	require.Equal(t, metricsviews.TimestampNotFound, result.Errors[0].Message)

	testutils.CreateModel(t, s, "AdBids_model", "select id, timestamp, publisher from AdBids", AdBidsModelRepoPath)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 2, 0, 1, 0, AdBidsDashboardAffectedPaths)
	require.Equal(t, `dimension not found: domain`, result.Errors[0].Message)
	require.Equal(t, []string{"Dimensions", "1"}, result.Errors[0].PropertyPath)
	require.Contains(t, result.Errors[1].Message, `Binder Error: Referenced column "bid_price" not found`)
	require.Equal(t, []string{"Measures", "1"}, result.Errors[1].PropertyPath)

	// ignore invalid measure and dimension
	time.Sleep(time.Millisecond * 10)
	err = s.Repo.Put(context.Background(), s.InstID, AdBidsDashboardRepoPath, strings.NewReader(`model: AdBids_model
timeseries: timestamp
timegrains:
- 1 day
- 1 month
default_timegrain: ""
dimensions:
- label: Publisher
  property: publisher
- label: Domain
  property: domain
  ignore: true
measures:
- expression: count(*)
- expression: avg(bid_price)
  ignore: true
`))
	require.NoError(t, err)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 1, 0, 0, []string{AdBidsDashboardRepoPath})
	mvEntry := testutils.AssertInCatalogStore(t, s, "AdBids_dashboard", AdBidsDashboardRepoPath)
	mv := mvEntry.GetMetricsView()
	require.Len(t, mv.Measures, 1)
	require.Equal(t, "count(*)", mv.Measures[0].Expression)
	require.Len(t, mv.Dimensions, 1)
	require.Equal(t, "publisher", mv.Dimensions[0].Name)

	time.Sleep(time.Millisecond * 10)
	err = s.Repo.Put(context.Background(), s.InstID, AdBidsDashboardRepoPath, strings.NewReader(`model: AdBids_model
timeseries: timestamp
timegrains:
- 1 day
- 1 month
default_timegrain: ""
dimensions:
- label: Publisher
  property: publisher
- label: Domain
  property: domain
  ignore: true
measures:
- expression: count(*)
  ignore: true
- expression: avg(bid_price)
  ignore: true
`))
	require.NoError(t, err)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 1, 0, 0, 0, []string{AdBidsDashboardRepoPath})
	require.Equal(t, metricsviews.MissingMeasure, result.Errors[0].Message)

	time.Sleep(time.Millisecond * 10)
	err = s.Repo.Put(context.Background(), s.InstID, AdBidsDashboardRepoPath, strings.NewReader(`model: AdBids_model
timeseries: timestamp
timegrains:
- 1 day
- 1 month
default_timegrain: ""
dimensions:
- label: Publisher
  property: publisher
  ignore: true
- label: Domain
  property: domain
  ignore: true
measures:
- expression: count(*)
- expression: avg(bid_price)
  ignore: true
`))
	require.NoError(t, err)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 1, 0, 0, 0, []string{AdBidsDashboardRepoPath})
	require.Equal(t, metricsviews.MissingDimension, result.Errors[0].Message)
}

func TestInvalidFiles(t *testing.T) {
	s, _ := initBasicService(t)
	ctx := context.Background()

	err := s.Repo.Put(ctx, s.InstID, AdBidsRepoPath, strings.NewReader(`type: local_file
path:
 - data/source.csv`))
	require.NoError(t, err)
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 3, 0, 0, 1, AdBidsAffectedPaths)
	require.Contains(t, result.Errors[0].Message, "yaml: unmarshal errors")

	testutils.CreateSource(t, s, "Ad-Bids", "AdBids.csv", "/sources/Ad-Bids.yaml")
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 1, 0, 0, 0, []string{"/sources/Ad-Bids.yaml"})
	require.Equal(t, "/sources/Ad-Bids.yaml", result.Errors[0].FilePath)
	require.Equal(t, "invalid file name", result.Errors[0].Message)
}

func TestReconcileDryRun(t *testing.T) {
	s, _ := initBasicService(t)

	AdBidsModelDashboardPath := []string{AdBidsModelRepoPath, AdBidsDashboardRepoPath}

	testutils.CreateModel(t, s, "AdBids_model", "select * from AdImpressions", AdBidsModelRepoPath)
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{
		DryRun: true,
	})
	require.NoError(t, err)
	// only one error returned. dashboard is still valid since model was not removed
	testutils.AssertMigration(t, result, 1, 0, 0, 0,
		[]string{AdBidsModelRepoPath})
	testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)
	testutils.AssertInCatalogStore(t, s, "AdBids_dashboard", AdBidsDashboardRepoPath)
	// commit the update
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 2, 0, 0, 0, AdBidsModelDashboardPath)

	// error should be returned after reconcile
	time.Sleep(time.Millisecond * 10)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{
		DryRun:       true,
		ChangedPaths: AdBidsModelDashboardPath,
		ForcedPaths:  AdBidsModelDashboardPath,
	})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 2, 0, 0, 0, AdBidsModelDashboardPath)

	testutils.CreateModel(t, s, "AdBids_model",
		"select id, timestamp, publisher, domain, bid_price from AdBids", AdBidsModelRepoPath)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{
		DryRun: true,
	})
	require.NoError(t, err)
	// error is still returned for dashboard since model is not updated in dry run
	testutils.AssertMigration(t, result, 1, 0, 0, 0,
		[]string{AdBidsDashboardRepoPath})
	require.Equal(t, AdBidsDashboardRepoPath, result.Errors[0].FilePath)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	testutils.AssertMigration(t, result, 0, 2, 0, 0, AdBidsModelDashboardPath)
}

func TestReconcileNewFile(t *testing.T) {
	s, _ := initBasicService(t)

	testutils.CreateSource(t, s, "AdImpressions", AdImpressionsCsvPath, AdImpressionsRepoPath)
	// reconcile with changed paths
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{
		ChangedPaths: []string{AdBidsRepoPath},
	})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 0, 0, 0, []string{})

	time.Sleep(time.Millisecond * 10)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{
		ChangedPaths: []string{AdImpressionsRepoPath},
	})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 1, 0, 0, []string{AdImpressionsRepoPath})
}

func initBasicService(t *testing.T) (*catalog.Service, string) {
	s, dir := getService(t)
	testutils.CreateSource(t, s, "AdBids", AdBidsCsvPath, AdBidsRepoPath)
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 1, 0, 0, []string{AdBidsRepoPath})
	testutils.AssertTable(t, s, "AdBids", AdBidsRepoPath)

	testutils.CreateModel(t, s, "AdBids_model",
		"select id, timestamp, publisher, domain, bid_price from AdBids", AdBidsModelRepoPath)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 1, 0, 0, []string{AdBidsModelRepoPath})
	testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)

	testutils.CreateMetricsView(t, s, &runtimev1.MetricsView{
		Name:          "AdBids_dashboard",
		Model:         "AdBids_model",
		TimeDimension: "timestamp",
		TimeGrains:    []string{"1 day", "1 month"},
		Dimensions: []*runtimev1.MetricsView_Dimension{
			{
				Name:  "publisher",
				Label: "Publisher",
			},
			{
				Name:  "domain",
				Label: "Domain",
			},
		},
		Measures: []*runtimev1.MetricsView_Measure{
			{
				Expression: "count(*)",
			},
			{
				Expression: "avg(bid_price)",
			},
		},
	}, AdBidsDashboardRepoPath)
	result, err = s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 1, 0, 0, []string{AdBidsDashboardRepoPath})
	testutils.AssertInCatalogStore(t, s, "AdBids_dashboard", AdBidsDashboardRepoPath)

	return s, dir
}

func getService(t *testing.T) (*catalog.Service, string) {
	dir := t.TempDir()

	duckdbStore, err := drivers.Open("duckdb", filepath.Join(dir, "stage.db"))
	require.NoError(t, err)
	err = duckdbStore.Migrate(context.Background())
	require.NoError(t, err)
	olap, ok := duckdbStore.OLAPStore()
	require.True(t, ok)
	catalogObject, ok := duckdbStore.CatalogStore()
	require.True(t, ok)

	fileStore, err := drivers.Open("file", dir)
	require.NoError(t, err)
	repo, ok := fileStore.RepoStore()
	require.True(t, ok)

	return catalog.NewService(catalogObject, repo, olap, "test", nil), dir
}
