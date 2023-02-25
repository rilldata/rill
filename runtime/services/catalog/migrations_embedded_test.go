package catalog_test

import (
	"context"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/services/catalog"
	"github.com/rilldata/rill/runtime/services/catalog/testutils"
	"github.com/stretchr/testify/require"
)

var EmbeddedSourceName = "a5679a659bbebf0ea9bf47a382e380b7b"
var EmbeddedSourcePath = "data/AdBids.csv"
var EmbeddedGzSourceName = "af7de1e5523cb411afc9c5112d5e2cc1e"
var EmbeddedGzSourcePath = "data/AdBids.csv.gz"
var ImpEmbeddedSourceName = "aee12850b230f8be0ff25d8d9d65648c7"
var ImpEmbeddedSourcePath = "data/AdImpressions.csv"
var AdBidsNewModeName = "AdBids_new_model"
var AdBidsNewModelPath = "/models/AdBids_new_model.sql"

func TestEmbeddedSourcesHappyPath(t *testing.T) {
	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsNewModelPath, AdBidsModelRepoPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			testutils.CopyFileToData(t, dir, AdBidsCsvPath, "AdBids.csv")

			addEmbeddedModel(t, s)
			addEmbeddedNewModel(t, s)
			testutils.AssertTable(t, s, "AdBids_new_model", AdBidsNewModelPath)

			result, err := s.Reconcile(context.Background(), tt.config)
			// no errors when reconcile is run later
			testutils.AssertMigration(t, result, 0, 0, 0, 0, []string{})
			require.NoError(t, err)

			// delete on of the models
			err = os.Remove(path.Join(dir, AdBidsNewModelPath))
			time.Sleep(10 * time.Millisecond)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 0, 1, 1, []string{AdBidsNewModelPath, EmbeddedSourcePath})
			testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)

			// delete the other model
			err = os.Remove(path.Join(dir, AdBidsModelRepoPath))
			time.Sleep(10 * time.Millisecond)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(
				t,
				result,
				1,
				0,
				0,
				2,
				[]string{AdBidsModelRepoPath, AdBidsDashboardRepoPath, EmbeddedSourcePath},
			)
			testutils.AssertTableAbsence(t, s, EmbeddedSourceName)
		})
	}
}

func TestEmbeddedSourcesQueryChanging(t *testing.T) {
	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsNewModelPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			testutils.CopyFileToData(t, dir, AdBidsCsvPath, "AdBids.csv")
			testutils.CopyFileToData(t, dir, AdBidsCsvGzPath, "AdBids.csv.gz")

			addEmbeddedModel(t, s)
			addEmbeddedNewModel(t, s)

			testutils.CreateModel(
				t,
				s,
				AdBidsNewModeName,
				`select id, timestamp, publisher from "data/AdBids.csv.gz"`,
				AdBidsNewModelPath,
			)
			result, err := s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 1, 2, 0, []string{AdBidsNewModelPath, EmbeddedSourcePath, EmbeddedGzSourcePath})
			adBidsEntry := testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)
			require.ElementsMatch(t, []string{"adbids_model"}, adBidsEntry.Children)
			adBidsGzEntry := testutils.AssertTable(t, s, EmbeddedGzSourceName, EmbeddedGzSourcePath)
			require.ElementsMatch(t, []string{strings.ToLower(AdBidsNewModeName)}, adBidsGzEntry.Children)
			testutils.AssertTable(t, s, AdBidsNewModeName, AdBidsNewModelPath)

			sc, _ := copyService(t, s)
			testutils.CreateModel(
				t,
				s,
				AdBidsNewModeName,
				`select id, timestamp, publisher from "data/AdBids.csv"`,
				AdBidsNewModelPath,
			)
			result, err = sc.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 0, 2, 1, []string{AdBidsNewModelPath, EmbeddedSourcePath, EmbeddedGzSourcePath})
			adBidsEntry = testutils.AssertTable(t, sc, EmbeddedSourceName, EmbeddedSourcePath)
			require.ElementsMatch(t, []string{"adbids_model", strings.ToLower(AdBidsNewModeName)}, adBidsEntry.Children)
			testutils.AssertTableAbsence(t, sc, EmbeddedGzSourceName)
		})
	}
}

func TestEmbeddedMultipleSources(t *testing.T) {
	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsNewModelPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			testutils.CopyFileToData(t, dir, AdBidsCsvPath, "AdBids.csv")
			testutils.CopyFileToData(t, dir, AdImpressionsCsvPath, "AdImpressions.csv")

			// create a model with 2 embedded sources, one repeated twice
			testutils.CreateModel(
				t,
				s,
				AdBidsNewModeName,
				`with
    NewYorkImpressions as (
        select imp.id, imp.city, imp.country from "data/AdImpressions.csv" imp where imp.city = 'NewYork'
    )
    select count(*) as impressions, bid.publisher, bid.domain, imp.city, imp.country
    from "data/AdBids.csv" bid join "data/AdImpressions.csv" imp on bid.id = imp.id
    group by bid.publisher, bid.domain, imp.city, imp.country`,
				AdBidsNewModelPath,
			)
			result, err := s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 3, 0, 0, []string{AdBidsNewModelPath, EmbeddedSourcePath, ImpEmbeddedSourcePath})
			adBidsEntry := testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)
			require.ElementsMatch(t, []string{strings.ToLower(AdBidsNewModeName)}, adBidsEntry.Children)
			adImpEntry := testutils.AssertTable(t, s, ImpEmbeddedSourceName, ImpEmbeddedSourcePath)
			require.ElementsMatch(t, []string{strings.ToLower(AdBidsNewModeName)}, adImpEntry.Children)
			modelEntry := testutils.AssertTable(t, s, AdBidsNewModeName, AdBidsNewModelPath)
			require.ElementsMatch(t, []string{strings.ToLower(EmbeddedSourceName), strings.ToLower(ImpEmbeddedSourceName)}, modelEntry.Parents)

			// update the model to have embedded sources without repetitions
			testutils.CreateModel(
				t,
				s,
				AdBidsNewModeName,
				`select count(*) as impressions, bid.publisher, bid.domain, imp.city, imp.country
    from "data/AdBids.csv" bid join "data/AdImpressions.csv" imp on bid.id = imp.id
    group by bid.publisher, bid.domain, imp.city, imp.country`,
				AdBidsNewModelPath,
			)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 0, 1, 0, []string{AdBidsNewModelPath})
			adBidsEntry = testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)
			require.ElementsMatch(t, []string{strings.ToLower(AdBidsNewModeName)}, adBidsEntry.Children)
			adImpEntry = testutils.AssertTable(t, s, ImpEmbeddedSourceName, ImpEmbeddedSourcePath)
			require.ElementsMatch(t, []string{strings.ToLower(AdBidsNewModeName)}, adImpEntry.Children)
			modelEntry = testutils.AssertTable(t, s, AdBidsNewModeName, AdBidsNewModelPath)
			require.ElementsMatch(t, []string{strings.ToLower(EmbeddedSourceName), strings.ToLower(ImpEmbeddedSourceName)}, modelEntry.Parents)
		})
	}
}

func TestEmbeddedSourceOnNewService(t *testing.T) {
	s, dir := initBasicService(t)

	testutils.CopyFileToData(t, dir, AdBidsCsvPath, "AdBids.csv")
	addEmbeddedModel(t, s)

	sc, result := copyService(t, s)
	// no updates other than when a new service is started
	// dashboards don't have equals check implemented right now. hence it is updated here
	testutils.AssertMigration(t, result, 0, 0, 2, 0, []string{AdBidsDashboardRepoPath, EmbeddedSourcePath})

	addEmbeddedNewModel(t, s)

	// change one model back to use AdBids from embedding the source
	testutils.CreateModel(
		t,
		sc,
		"AdBids_model",
		`select id, timestamp, publisher, domain, bid_price from AdBids`,
		AdBidsModelRepoPath,
	)
	// delete the other model embedding the source
	err := os.Remove(path.Join(dir, AdBidsNewModelPath))
	require.NoError(t, err)
	// create another copy
	sc, result = copyService(t, s)
	testutils.AssertMigration(
		t,
		result,
		0,
		0,
		2,
		2,
		[]string{EmbeddedSourcePath, AdBidsModelRepoPath, AdBidsDashboardRepoPath, AdBidsNewModelPath},
	)
	testutils.AssertTableAbsence(t, s, EmbeddedSourceName)
}

func TestEmbeddingModelRename(t *testing.T) {
	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsModelRepoPath, AdBidsNewModelPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			testutils.CopyFileToData(t, dir, AdBidsCsvPath, "AdBids.csv")
			addEmbeddedModel(t, s)

			testutils.RenameFile(t, dir, AdBidsModelRepoPath, AdBidsNewModelPath)
			result, err := s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(
				t,
				result,
				1,
				0,
				2,
				0,
				[]string{AdBidsDashboardRepoPath, EmbeddedSourcePath, AdBidsNewModelPath},
			)
			adBidsEntry := testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)
			require.ElementsMatch(t, []string{strings.ToLower(AdBidsNewModeName)}, adBidsEntry.Children)
		})
	}
}

func TestEmbeddedSourceRefresh(t *testing.T) {
	s, dir := initBasicService(t)

	testutils.CopyFileToData(t, dir, AdBidsCsvPath, "AdBids.csv")
	addEmbeddedModel(t, s)

	testutils.CopyFileToData(t, dir, AdImpressionsCsvPath, "AdBids.csv")
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{
		ChangedPaths: []string{EmbeddedSourcePath},
		ForcedPaths:  []string{EmbeddedSourcePath},
	})
	require.NoError(t, err)
	// refreshing the embedded source and replacing with different file caused errors
	// the model depended on column not present in the new file
	testutils.AssertMigration(
		t,
		result,
		2,
		0,
		1,
		0,
		[]string{EmbeddedSourcePath, AdBidsModelRepoPath, AdBidsDashboardRepoPath},
	)
}

func TestEmbeddedSourcesErroredOut(t *testing.T) {
	configs := []struct {
		title  string
		config catalog.ReconcileConfig
	}{
		{"ReconcileAll", catalog.ReconcileConfig{}},
		{"ReconcileSelected", catalog.ReconcileConfig{
			ChangedPaths: []string{AdBidsNewModelPath},
		}},
	}

	for _, tt := range configs {
		t.Run(tt.title, func(t *testing.T) {
			s, dir := initBasicService(t)

			testutils.CopyFileToData(t, dir, AdBidsCsvPath, "AdBids.csv")
			addEmbeddedModel(t, s)
			addEmbeddedNewModel(t, s)

			// change the model to point to invalid file
			testutils.CreateModel(
				t,
				s,
				"AdBids_model",
				`select id, timestamp, publisher, domain, bid_price from "data/AdBids.cs"`,
				AdBidsNewModelPath,
			)
			result, err := s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 2, 1, 1, 0, []string{AdBidsNewModelPath, EmbeddedSourcePath, "data/AdBids.cs"})
			adBidsEntry := testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)
			require.ElementsMatch(t, []string{strings.ToLower("AdBids_model")}, adBidsEntry.Children)

			// change back to original valid file
			testutils.CreateModel(
				t,
				s,
				"AdBids_model",
				`select id, timestamp, publisher, domain, bid_price from "data/AdBids.csv"`,
				AdBidsNewModelPath,
			)
			result, err = s.Reconcile(context.Background(), tt.config)
			require.NoError(t, err)
			testutils.AssertMigration(t, result, 0, 1, 1, 0, []string{AdBidsNewModelPath, EmbeddedSourcePath})
			adBidsEntry = testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)
			require.ElementsMatch(t, []string{strings.ToLower("AdBids_model"), strings.ToLower(AdBidsNewModeName)}, adBidsEntry.Children)
		})
	}
}

func addEmbeddedModel(t *testing.T, s *catalog.Service) {
	testutils.CreateModel(
		t,
		s,
		"AdBids_model",
		`select id, timestamp, publisher, domain, bid_price from "data/AdBids.csv"`,
		AdBidsModelRepoPath,
	)
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 1, 2, 0, []string{AdBidsModelRepoPath, AdBidsDashboardRepoPath, EmbeddedSourcePath})
	testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)
	testutils.AssertTable(t, s, "AdBids_model", AdBidsModelRepoPath)
}

func addEmbeddedNewModel(t *testing.T, s *catalog.Service) {
	testutils.CreateModel(
		t,
		s,
		AdBidsNewModeName,
		`select id, timestamp, publisher from "data/AdBids.csv"`,
		AdBidsNewModelPath,
	)
	result, err := s.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	testutils.AssertMigration(t, result, 0, 1, 1, 0, []string{AdBidsNewModelPath, EmbeddedSourcePath})
	testutils.AssertTable(t, s, EmbeddedSourceName, EmbeddedSourcePath)
	testutils.AssertTable(t, s, AdBidsNewModeName, AdBidsNewModelPath)
}

func copyService(t *testing.T, s *catalog.Service) (*catalog.Service, *catalog.ReconcileResult) {
	sc := catalog.NewService(s.Catalog, s.Repo, s.Olap, s.RegistryStore, s.InstID, nil)
	result, err := sc.Reconcile(context.Background(), catalog.ReconcileConfig{})
	require.NoError(t, err)
	return sc, result
}
