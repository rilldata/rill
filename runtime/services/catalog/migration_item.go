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
	Type                   MigrationType
	CatalogInFile          *drivers.CatalogEntry
	CatalogInStore         *drivers.CatalogEntry
	NewCatalog             *drivers.CatalogEntry
	HasChanged             bool
	FromName               string
	FromNormalizedName     string
	FromPath               string
	NormalizedDependencies []string
	Error                  *runtimev1.ReconcileError
}

func (i *MigrationItem) renameFrom(name, path string) {
	i.Type = MigrationRename
	i.FromName = name
	i.FromNormalizedName = normalizeName(name)
	i.FromPath = path
}

type MigrationType int

const (
	MigrationNoChange     MigrationType = 0
	MigrationCreate       MigrationType = 1
	MigrationRename       MigrationType = 2
	MigrationUpdate       MigrationType = 3
	MigrationReportUpdate MigrationType = 4
	MigrationDelete       MigrationType = 5
)

func (s *Service) getMigrationItems(
	ctx context.Context,
	repoPath string,
	storeObjectsMap map[string]*drivers.CatalogEntry,
	forcedPathMap map[string]bool,
) []*MigrationItem {
	var items []*MigrationItem
	// primary item for repoPath
	var item *MigrationItem
	hasFileObject := false

	catalog, err := artifacts.Read(ctx, s.Repo, s.InstID, repoPath)
	if err != nil {
		item = s.newMigrationItemFromError(repoPath, err)
		items = []*MigrationItem{item}
	} else {
		item, items = s.newMigrationItemFromFile(ctx, repoPath, catalog, storeObjectsMap)
		items = append(items, item)
		hasFileObject = true
	}
	// mark the item as changed if it was forced to update
	item.HasChanged = forcedPathMap[repoPath]

	if hasFileObject {
		s.checkFileChange(ctx, item)
	}

	catalogInStore, hasStoreObject := storeObjectsMap[item.NormalizedName]
	if hasStoreObject {
		item.CatalogInStore = catalogInStore
		if !hasFileObject {
			item.NewCatalog = catalogInStore
		}
	}

	if item.Type == MigrationNoChange && item.HasChanged {
		if hasStoreObject {
			item.Type = MigrationUpdate
		} else {
			item.Type = MigrationCreate
		}
	}

	if !hasFileObject && !hasStoreObject {
		// invalid file created/updated
		// do not run any migration on this, changed are already saved to the file
		item.Type = MigrationNoChange
		if errors.Is(err, artifacts.ErrFileRead) {
			// the item is possibly for a file that doesn't exist but was passed in ChangedPaths
			return nil
		}
	}

	if hasFileObject && hasStoreObject {
		if item.Name != catalogInStore.Name {
			// rename with same name different case
			item.renameFrom(catalogInStore.Name, item.Path)
		}

		switch item.Type {
		case MigrationCreate:
			if migrator.IsEqual(ctx, item.CatalogInFile, item.CatalogInStore) && !item.HasChanged {
				// if the actual content has not changed, mark as MigrationNoChange
				item.Type = MigrationNoChange
			} else {
				// else mark as MigrationUpdate
				item.Type = MigrationUpdate
			}

		case MigrationNoChange:
			// if item doesn't exist in olap, mark as create
			// happens when the catalog table is modified directly in some way
			ok, _ := migrator.ExistsInOlap(ctx, s.Olap, item.CatalogInFile)
			if !ok {
				item.Type = MigrationCreate
			}
		}
	}

	return items
}

func (s *Service) newMigrationItemFromError(repoPath string, err error) *MigrationItem {
	item := &MigrationItem{
		Type: MigrationNoChange,
		Path: repoPath,
	}

	if !errors.Is(err, artifacts.ErrFileRead) {
		item.Error = &runtimev1.ReconcileError{
			Code:     runtimev1.ReconcileError_CODE_SYNTAX,
			Message:  err.Error(),
			FilePath: repoPath,
		}
	}
	item.Type = MigrationDelete

	item.Name = fileutil.Stem(repoPath)
	item.NormalizedName = normalizeName(item.Name)
	return item
}

func (s *Service) newMigrationItemFromFile(
	ctx context.Context,
	repoPath string,
	catalogInFile *drivers.CatalogEntry,
	storeObjectsMap map[string]*drivers.CatalogEntry,
) (*MigrationItem, []*MigrationItem) {
	item := &MigrationItem{
		Type:           MigrationNoChange,
		Path:           repoPath,
		Name:           catalogInFile.Name,
		NormalizedName: normalizeName(catalogInFile.Name),
		CatalogInFile:  catalogInFile,
		NewCatalog:     catalogInFile,
	}
	item.Name = catalogInFile.Name
	item.NormalizedName = normalizeName(item.Name)
	item.CatalogInFile = catalogInFile

	normalizedDependencies, embeddedEntries := migrator.GetDependencies(ctx, s.Olap, item.CatalogInFile)
	// convert dependencies to lower case
	for i, dep := range normalizedDependencies {
		normalizedDependencies[i] = normalizeName(dep)
	}
	item.NormalizedDependencies = normalizedDependencies

	return item, s.resolveDependencies(item, storeObjectsMap, embeddedEntries)
}

func (s *Service) checkFileChange(ctx context.Context, item *MigrationItem) {
	repoStat, _ := s.Repo.Stat(ctx, s.InstID, item.Path)
	catalogLastUpdated, _ := migrator.LastUpdated(ctx, s.InstID, s.Repo, item.CatalogInFile)
	if repoStat.LastUpdated.After(catalogLastUpdated) {
		item.CatalogInFile.UpdatedOn = repoStat.LastUpdated
	} else {
		item.CatalogInFile.UpdatedOn = catalogLastUpdated
		// if catalog has changed in any way then always re-create/update
		item.HasChanged = true
	}
	if item.CatalogInFile.UpdatedOn.After(s.LastMigration) {
		// assume creation until we see a catalog object
		item.Type = MigrationCreate
	}
}

func (s *Service) resolveDependencies(
	item *MigrationItem,
	storeObjectsMap map[string]*drivers.CatalogEntry,
	embeddedEntries []*drivers.CatalogEntry,
) []*MigrationItem {
	items := make([]*MigrationItem, 0)

	prevEmbeddedEntries := make(map[string]bool)
	prevDependencies := s.dag.GetParents(item.NormalizedName)
	for _, prevDependency := range prevDependencies {
		prevEmbeddedEntries[prevDependency] = true
	}

	for _, embeddedEntry := range embeddedEntries {
		normalizedEmbeddedName := normalizeName(embeddedEntry.Name)
		if prevEmbeddedEntries[normalizedEmbeddedName] {
			// delete from map for unchanged embedded entry.
			// this map will later be used to remove link from previously embedded entry
			delete(prevEmbeddedEntries, normalizedEmbeddedName)
			continue
		}
		embeddedItem := s.newEmbeddedMigrationItem(embeddedEntry, MigrationCreate)
		if existingEntry, ok := storeObjectsMap[embeddedItem.NormalizedName]; ok {
			// update the catalog for embedded entry to the one from store
			embeddedItem.CatalogInFile = existingEntry
			embeddedItem.CatalogInStore = existingEntry
			embeddedItem.Type = MigrationReportUpdate
		}
		items = append(items, embeddedItem)
	}

	// go through previous embedded entries not embedded anymore
	for prevEmbeddedEntry := range prevEmbeddedEntries {
		existingEntry, ok := storeObjectsMap[prevEmbeddedEntry]
		if !ok || !existingEntry.Embedded {
			// should not happen
			continue
		}
		embeddedItem := s.newEmbeddedMigrationItem(existingEntry, MigrationReportUpdate)
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
		NewCatalog:     newEntry,
		Type:           migrationType,
	}
}

func (s *Service) newDeleteMigrationItem(entry *drivers.CatalogEntry) *MigrationItem {
	return &MigrationItem{
		Name:           entry.Name,
		NormalizedName: normalizeName(entry.Name),
		Type:           MigrationDelete,
		Path:           entry.Path,
		CatalogInStore: entry,
		NewCatalog:     entry,
	}
}

func normalizeName(name string) string {
	return strings.ToLower(name)
}
