package catalog

import (
	"context"
	"errors"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/arrayutil"
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
	embeddedMigrations map[string]*MigrationItem,
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

		items = append(items, s.resolveDependencies(ctx, item, embeddedEntries, embeddedMigrations)...)

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

	catalogInStore, ok := storeObjectsMap[item.NormalizedName]

	if item.Type == MigrationNoChange && forcedPathMap[repoPath] {
		if ok {
			item.Type = MigrationUpdate
		} else {
			item.Type = MigrationCreate
		}
	}

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
		item.renameFrom(catalogInStore.Name, item.Path)
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
	embeddedMigrations map[string]*MigrationItem,
) []*MigrationItem {
	items := make([]*MigrationItem, 0)

	prevEmbeddedEntries := make(map[string]bool)
	prevDependencies := s.dag.GetParents(item.NormalizedName)
	for _, prevDependency := range prevDependencies {
		prevEmbeddedEntries[prevDependency] = true
	}
	for _, prevEmbedded := range item.CatalogInFile.Embeds {
		prevEmbeddedEntries[prevEmbedded] = true
	}

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
		embeddedItem, ok := embeddedMigrations[normalizedEmbeddedName]
		if !ok {
			embeddedItem = s.newEmbeddedMigrationItem(embeddedEntry, MigrationCreate)
			if existingEntry, ok := s.Catalog.FindEntry(ctx, s.InstID, embeddedEntry.Name); ok {
				// update the catalog for embedded entry to the one from store
				embeddedItem.CatalogInFile = existingEntry
				embeddedItem.CatalogInStore = existingEntry
				if arrayutil.Contains(existingEntry.Embeds, item.NormalizedName) {
					// if it already has this, no change
					embeddedItem.Type = MigrationNoChange
				} else {
					// else mark as catalog update, this means
					embeddedItem.Type = MigrationUpdateCatalog
				}
			}
		}
		embeddedItem.addLink(item.NormalizedName)
		items = append(items, embeddedItem)
	}

	// go through previous embedded entries not embedded anymore
	for prevEmbeddedEntry := range prevEmbeddedEntries {
		embeddedItem, ok := embeddedMigrations[prevEmbeddedEntry]
		if !ok {
			existingEntry, ok := s.Catalog.FindEntry(ctx, s.InstID, prevEmbeddedEntry)
			if !ok || !existingEntry.Embedded {
				continue
			}
			embeddedItem = s.newEmbeddedMigrationItem(existingEntry, MigrationUpdateCatalog)
		}
		embeddedItem.removeLink(item.NormalizedName)
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

func (i *MigrationItem) addLink(name string) {
	if arrayutil.Contains(i.CatalogInFile.Embeds, name) {
		return
	}
	if i.Type == MigrationNoChange {
		i.Type = MigrationUpdateCatalog
	}
	i.CatalogInFile.Links++
	i.CatalogInFile.Embeds = append(i.CatalogInFile.Embeds, name)
}

func (i *MigrationItem) removeLink(name string) {
	i.CatalogInFile.Links--
	i.CatalogInFile.Embeds = arrayutil.Delete(i.CatalogInFile.Embeds, name)
	if i.CatalogInFile.Links == 0 {
		i.Type = MigrationDelete
	} else if i.Type == MigrationNoChange {
		i.Type = MigrationUpdateCatalog
	}
}

func normalizeName(name string) string {
	return strings.ToLower(name)
}
