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
	Type                   MigrationType
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

type MigrationType int

const (
	MigrationNoChange      MigrationType = 0
	MigrationCreate        MigrationType = 1
	MigrationRename        MigrationType = 2
	MigrationUpdate        MigrationType = 3
	MigrationUpdateCatalog MigrationType = 4
	MigrationDelete        MigrationType = 5
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
	var embeddedEntries []*drivers.CatalogEntry

	catalog, err := artifacts.Read(ctx, s.Repo, s.InstID, repoPath)
	if err != nil {
		if !errors.Is(err, artifacts.ErrFileRead) {
			item.Error = &runtimev1.ReconcileError{
				Code:     runtimev1.ReconcileError_CODE_SYNTAX,
				Message:  err.Error(),
				FilePath: repoPath,
			}
		}

		item.Name = fileutil.Stem(repoPath)
		item.NormalizedName = normalizeName(item.Name)
		item.Type = MigrationDelete
	} else {
		item.Name = catalog.Name
		item.NormalizedName = normalizeName(item.Name)
		item.CatalogInFile = catalog

		var normalizedDependencies []string
		normalizedDependencies, embeddedEntries = migrator.GetDependencies(ctx, s.Olap, item.CatalogInFile)
		// convert dependencies to lower case
		for i, dep := range normalizedDependencies {
			normalizedDependencies[i] = normalizeName(dep)
		}
		item.NormalizedDependencies = normalizedDependencies

		items = append(items, s.resolveDependencies(ctx, item, embeddedEntries)...)

		repoStat, _ := s.Repo.Stat(ctx, s.InstID, repoPath)
		catalogLastUpdated, _ := migrator.LastUpdated(ctx, s.InstID, s.Repo, catalog)
		if repoStat.LastUpdated.After(catalogLastUpdated) {
			item.CatalogInFile.UpdatedOn = repoStat.LastUpdated
		} else {
			item.CatalogInFile.UpdatedOn = catalogLastUpdated
			// if catalog has changed in any way then always re-create/update
			forceChange = true
		}
		if item.CatalogInFile.UpdatedOn.After(s.LastMigration) {
			// assume creation until we see a catalog object
			item.Type = MigrationCreate
		}
	}

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

func (s *Service) resolveDependencies(
	ctx context.Context,
	item *MigrationItem,
	embeddedEntries []*drivers.CatalogEntry,
) []*MigrationItem {
	items := make([]*MigrationItem, 0)

	prevEmbeddedEntries := make(map[string]bool)
	prevDependencies := s.dag.GetParents(item.NormalizedName)
	for _, prevDependency := range prevDependencies {
		prevEmbeddedEntries[prevDependency] = true
	}
	// TODO: handle 1st time run

	item.CatalogInFile.Embeds = make([]string, 0)

	for _, embeddedEntry := range embeddedEntries {
		normalizedEmbeddedName := normalizeName(embeddedEntry.Name)
		item.CatalogInFile.Embeds = append(item.CatalogInFile.Embeds, normalizedEmbeddedName)
		if prevEmbeddedEntries[normalizedEmbeddedName] {
			// delete from map for unchanged embedded entry.
			// this map will later be used to remove link from previously embedded entry
			delete(prevEmbeddedEntries, normalizedEmbeddedName)
			continue
		}
		embeddedItem := s.newEmbeddedMigrationItem(embeddedEntry, MigrationCreate)
		if existingEntry, ok := s.Catalog.FindEntry(ctx, s.InstID, embeddedEntry.Name); ok {
			existingEntry.Links++
			embeddedItem.CatalogInFile = existingEntry
			embeddedItem.Type = MigrationUpdateCatalog
		}
		items = append(items, embeddedItem)
	}

	for prevEmbeddedEntry := range prevEmbeddedEntries {
		existingEntry, ok := s.Catalog.FindEntry(ctx, s.InstID, prevEmbeddedEntry)
		if !ok || !existingEntry.Embedded {
			continue
		}
		existingEntry.Links--
		embeddedItem := s.newEmbeddedMigrationItem(existingEntry, MigrationUpdate)
		if existingEntry.Links == 0 {
			embeddedItem.Type = MigrationDelete
		}
		items = append(items, embeddedItem)
	}

	return items
}

func (s *Service) newEmbeddedMigrationItem(newEntry *drivers.CatalogEntry, migrationType MigrationType) *MigrationItem {
	return &MigrationItem{
		Name:           newEntry.Name,
		NormalizedName: strings.ToLower(newEntry.Name),
		Path:           newEntry.Path,
		CatalogInFile:  newEntry,
		CatalogInStore: newEntry,
		Type:           migrationType,
	}
}

func normalizeName(name string) string {
	return strings.ToLower(name)
}
