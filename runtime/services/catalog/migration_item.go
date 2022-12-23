package catalog

import (
	"context"
	"errors"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
)

type MigrationItem struct {
	Name                   string
	NormalizedName         string
	Path                   string
	CatalogInFile          *drivers.CatalogEntry
	CatalogInStore         *drivers.CatalogEntry
	Type                   int
	FromName               string
	FromPath               string
	NormalizedDependencies []string
	Error                  *runtimev1.ReconcileError
}

func (i *MigrationItem) renameFrom(from *MigrationItem) {
	i.Type = MigrationRename
	i.FromName = from.Name
	i.FromPath = from.Path
}

const (
	MigrationNoChange    int = 0
	MigrationCreate      int = 1
	MigrationRename      int = 2
	MigrationUpdate      int = 3
	MigrationDelete      int = 4
	MigrationNewArtifact int = 5
)

func (s *Service) getMigrationItem(
	ctx context.Context,
	repoPath string,
	storeObjectsMap map[string]*drivers.CatalogEntry,
	forcedPathMap map[string]bool,
) []*MigrationItem {
	item := &MigrationItem{
		Type: MigrationNoChange,
		Path: repoPath,
	}

	forceChange := forcedPathMap[repoPath]
	items := []*MigrationItem{item}

	catalog, err := artifacts.Read(ctx, s.Repo, s.InstID, repoPath)
	if err != nil {
		if !errors.Is(err, artifacts.ErrFileRead) {
			item.Error = &runtimev1.ReconcileError{
				Code:     runtimev1.ReconcileError_CODE_SYNTAX,
				Message:  err.Error(),
				FilePath: repoPath,
			}
		}
		name, ok := s.PathToName[repoPath]
		if ok {
			item.Name = name
		} else {
			item.Name = fileutil.Stem(repoPath)
		}

		item.Type = MigrationDelete
	} else {
		item.Name = catalog.Name
		item.CatalogInFile = catalog

		normalizedDependencies, newEntries := migrator.GetDependencies(ctx, s.Olap, catalog)
		for _, newEntry := range newEntries {
			if _, ok := s.Catalog.FindEntry(ctx, s.InstID, newEntry.Name); ok {
				// already exists
				// TODO: update links
				continue
			}
			// TODO: mark as embedded in artifact
			items = append(items, s.newEmbeddedMigrationItem(newEntry))
		}

		// convert dependencies to lower case
		for i, dep := range normalizedDependencies {
			normalizedDependencies[i] = strings.ToLower(dep)
		}
		item.NormalizedDependencies = normalizedDependencies

		repoStat, _ := s.Repo.Stat(ctx, s.InstID, repoPath)
		catalogLastUpdated, _ := migrator.LastUpdated(ctx, s.InstID, s.Repo, catalog)
		if repoStat.LastUpdated.After(catalogLastUpdated) {
			item.CatalogInFile.UpdatedOn = repoStat.LastUpdated
		} else {
			item.CatalogInFile.UpdatedOn = catalogLastUpdated
			// if catalog has changed in anyway then always re-create/update
			forceChange = true
		}
		if item.CatalogInFile.UpdatedOn.After(s.LastMigration) {
			// assume creation until we see a catalog object
			item.Type = MigrationCreate
		}
	}
	item.NormalizedName = strings.ToLower(item.Name)

	if item.Type == MigrationNoChange && forcedPathMap[repoPath] {
		item.Type = MigrationUpdate
	}

	catalogInStore, ok := storeObjectsMap[item.NormalizedName]
	if !ok {
		if item.CatalogInFile == nil {
			item.Type = MigrationNoChange
			if errors.Is(err, artifacts.ErrFileRead) {
				// the item is possibly for a file that doesn't exist but was passed in ChangedPaths
				return nil
			}
		}
		return items
	}
	item.CatalogInStore = catalogInStore
	if item.Name != catalogInStore.Name && item.CatalogInFile != nil {
		// rename with same name different case
		item.FromName = catalogInStore.Name
		item.Type = MigrationRename
	}

	switch item.Type {
	case MigrationCreate:
		if migrator.IsEqual(ctx, item.CatalogInFile, item.CatalogInStore) && !forceChange {
			// if the actual content has not changed, mark as MigrationNoChange
			item.Type = MigrationNoChange
		} else {
			// else mark as MigrationUpdate
			item.Type = MigrationUpdate
		}

	case MigrationNoChange:
		// if item doesn't exist in olap, mark as create
		// happens when the catalog table is modified directly
		ok, _ := migrator.ExistsInOlap(ctx, s.Olap, item.CatalogInFile)
		if !ok {
			item.Type = MigrationCreate
		}
	}

	return items
}

func (s *Service) newEmbeddedMigrationItem(newEntry *drivers.CatalogEntry) *MigrationItem {
	return &MigrationItem{
		Name:           newEntry.Name,
		NormalizedName: strings.ToLower(newEntry.Name),
		Path:           newEntry.Path,
		CatalogInFile:  newEntry,
		Type:           MigrationNewArtifact,
	}
}
