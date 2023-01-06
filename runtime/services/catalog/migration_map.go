package catalog

import (
	"context"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/arrayutil"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
)

// getMigrationMap returns a map of string to MigrationItem for all paths or selected paths.
func (s *Service) getMigrationMap(ctx context.Context, conf ReconcileConfig) (map[string]*MigrationItem, []*runtimev1.ReconcileError, error) {
	// TODO: if the repo folder is source controlled we should leverage it to find changes
	// TODO: ListRecursive needs some kind of cache or optimisation
	repoPaths := conf.ChangedPaths
	changedPathsHint := len(conf.ChangedPaths) > 0
	changedPathsMap := make(map[string]bool)
	if changedPathsHint {
		for _, changedPath := range conf.ChangedPaths {
			changedPathsMap[changedPath] = true
		}
	} else {
		var err error
		repoPaths, err = s.Repo.ListRecursive(ctx, s.InstID, "{sources,models,dashboards}/*.{sql,yaml,yml}")
		if err != nil {
			return nil, nil, err
		}
	}

	forcedPathMap := make(map[string]bool)
	for _, forcedPath := range conf.ForcedPaths {
		forcedPathMap[forcedPath] = true
	}

	storeObjectsMap := make(map[string]*drivers.CatalogEntry)
	storeObjectsPathMap := make(map[string]*drivers.CatalogEntry)
	storeObjectsConsumed := make(map[string]bool)
	storeObjects := s.Catalog.FindEntries(ctx, s.InstID, drivers.ObjectTypeUnspecified)
	for _, storeObject := range storeObjects {
		storeObjectsMap[strings.ToLower(storeObject.Name)] = storeObject
		storeObjectsPathMap[storeObject.Path] = storeObject
	}

	migrationMap := make(map[string]*MigrationItem)
	embeddedMigrations := make(map[string]*MigrationItem)
	reconcileErrors := make([]*runtimev1.ReconcileError, 0)
	deletions := make(map[string]*MigrationItem)
	additions := make(map[string]*MigrationItem)

	for _, repoPath := range repoPaths {
		items := s.getMigrationItem(ctx, repoPath, storeObjectsMap, storeObjectsPathMap, forcedPathMap, embeddedMigrations)
		for _, item := range items {
			keepNew, errPath := s.isInvalidDuplicate(migrationMap, changedPathsHint, changedPathsMap, item)
			if errPath != "" {
				reconcileErrors = append(reconcileErrors, &runtimev1.ReconcileError{
					Code:     runtimev1.ReconcileError_CODE_UNSPECIFIED,
					Message:  "item with same name exists",
					FilePath: errPath,
				})
			}
			if !keepNew {
				continue
			}
			if item.CatalogInFile != nil && item.CatalogInFile.Embedded {
				embeddedMigrations[item.NormalizedName] = item
			}

			if s.lookForRenames(ctx, item, migrationMap, additions, deletions) {
				migrationMap[item.NormalizedName] = item
			}
			storeObjectsConsumed[item.NormalizedName] = true

			if !changedPathsHint {
				continue
			}
			// go through the children only if forced paths is false
			children := s.dag.GetChildren(item.NormalizedName)
			for _, child := range children {
				childPath, ok := s.NameToPath[child]
				if !ok || (changedPathsHint && changedPathsMap[childPath]) {
					// if there is no entry for name to path or already in forced path then ignore the child
					continue
				}

				childItems := s.getMigrationItem(ctx, childPath, storeObjectsMap, storeObjectsPathMap, forcedPathMap, embeddedMigrations)
				for _, childItem := range childItems {
					migrationMap[childItem.NormalizedName] = childItem
				}
			}
		}
	}

	embeddedStoreObjects := make([]*drivers.CatalogEntry, 0)

	for _, storeObject := range storeObjectsMap {
		lowerStoreName := strings.ToLower(storeObject.Name)
		// ignore consumed store objects
		if storeObjectsConsumed[lowerStoreName] ||
			// ignore tables and unspecified objects
			storeObject.Type == drivers.ObjectTypeTable || storeObject.Type == drivers.ObjectTypeUnspecified {
			continue
		}
		// if repo paths were forced and the catalog was not in the paths then ignore
		if _, ok := changedPathsMap[storeObject.Path]; changedPathsHint && !ok {
			continue
		}
		// ignore embedded sources
		if storeObject.Embedded {
			if !s.hasMigrated {
				// only add if it was the 1st migration
				embeddedStoreObjects = append(embeddedStoreObjects, storeObject)
			}
			continue
		}
		found := false
		// find any additions that match and mark it as a MigrationRename
		for _, addition := range additions {
			if migrator.IsEqual(ctx, addition.CatalogInFile, storeObject) {
				addition.renameFrom(storeObject.Name, storeObject.Path)
				delete(additions, addition.NormalizedName)
				found = true
				break
			}
		}
		// if no matching item is found, add as a MigrationDelete
		if !found {
			migrationMap[lowerStoreName] = &MigrationItem{
				Name:           storeObject.Name,
				NormalizedName: lowerStoreName,
				Type:           MigrationDelete,
				Path:           storeObject.Path,
				CatalogInStore: storeObject,
			}
		}
	}

	s.checkEmbeddedEntries(ctx, migrationMap, embeddedStoreObjects)

	return migrationMap, reconcileErrors, nil
}

// isInvalidDuplicate checks if one of the existing or a new item is invalid duplicate.
func (s *Service) isInvalidDuplicate(
	migrationMap map[string]*MigrationItem,
	changedPathsHint bool,
	changedPathsMap map[string]bool,
	item *MigrationItem,
) (bool, string) {
	errPath := ""

	existing, ok := migrationMap[item.NormalizedName]

	if ok {
		if item.CatalogInStore != nil && item.CatalogInStore.Embedded {
			// ignore duplicate check for embedded items
			return false, ""
		}

		keepNew := false
		if existing.Name != item.Name {
			// where it is a MigrationRename with different case
			// keep the one marked as rename
			if item.Type == MigrationRename {
				keepNew = true
			}
		} else {
			// if existing item was deleted
			if existing.Type == MigrationDelete ||
				// or if the existing has error whereas new one doest
				(item.Error != nil && existing.Error != nil) ||
				// or if the existing file was updated after new (this makes it so that the old one will be retained)
				(item.Error == nil && item.CatalogInFile != nil && existing.CatalogInFile != nil &&
					existing.CatalogInFile.UpdatedOn.After(item.CatalogInFile.UpdatedOn)) {
				// replace the existing with new
				keepNew = true
				errPath = existing.Path
			} else {
				errPath = item.Path
			}
		}
		return keepNew, errPath
	}

	if changedPathsHint {
		// this handles the case where an item is renamed with the same name but different case.
		if existingPath, ok := s.NameToPath[item.NormalizedName]; ok && existingPath != item.Path && !changedPathsMap[existingPath] {
			return false, item.Path
		}
	}

	return true, errPath
}

func (s *Service) lookForRenames(
	ctx context.Context,
	item *MigrationItem,
	migrationMap map[string]*MigrationItem,
	additions map[string]*MigrationItem,
	deletions map[string]*MigrationItem,
) bool {
	add := true

	switch item.Type {
	case MigrationCreate:
		// if item is created compare with deletions to look for renames
		found := false
		for _, deletion := range deletions {
			if migrator.IsEqual(ctx, item.CatalogInFile, deletion.CatalogInStore) {
				item.renameFrom(deletion.Name, deletion.Path)
				delete(deletions, deletion.NormalizedName)
				delete(migrationMap, deletion.NormalizedName)
				found = true
				break
			}
		}
		if !found {
			additions[item.NormalizedName] = item
		}

	case MigrationDelete:
		found := false
		// if item is deleted compare with additions to look for renames
		for _, addition := range additions {
			if item.CatalogInStore != nil && migrator.IsEqual(ctx, addition.CatalogInFile, item.CatalogInStore) {
				addition.renameFrom(item.Name, item.Path)
				delete(additions, addition.NormalizedName)
				add = false
				found = true
				break
			}
		}
		if !found {
			deletions[item.NormalizedName] = item
		}
	}

	return add
}

func (s *Service) checkEmbeddedEntries(
	ctx context.Context,
	migrationMap map[string]*MigrationItem,
	embeddedStoreObjects []*drivers.CatalogEntry,
) {
	// update embedded items
	for _, item := range migrationMap {
		if item.CatalogInStore == nil {
			if item.Type == MigrationRename {
				s.checkEmbeddingOnRename(item, migrationMap)
			}
			continue
		}
		for _, embedded := range item.CatalogInStore.Embeds {
			if item.CatalogInStore.Embedded {
				s.checkEmbeddedSourceEntry(migrationMap, embedded, item)
			} else {
				s.checkEmbeddedEntry(ctx, migrationMap, embedded, item)
			}
		}
	}

	for _, embeddedStoreObject := range embeddedStoreObjects {
		item := s.newEmbeddedMigrationItem(embeddedStoreObject, MigrationNoChange)
		for _, embedded := range embeddedStoreObject.Embeds {
			s.checkEmbeddedSourceEntry(migrationMap, embedded, item)
		}
	}
}

func (s *Service) checkEmbeddedEntry(
	ctx context.Context,
	migrationMap map[string]*MigrationItem,
	embedded string,
	item *MigrationItem,
) {
	embeddedMigrationItem, ok := migrationMap[embedded]
	if !ok {
		existingEntry, ok := s.Catalog.FindEntry(ctx, s.InstID, embedded)
		if !ok {
			return
		}
		embeddedMigrationItem = s.newEmbeddedMigrationItem(existingEntry, MigrationUpdateCatalog)
	}

	contains := arrayutil.Contains(embeddedMigrationItem.CatalogInFile.Embeds, item.NormalizedName)
	if item.Type == MigrationDelete && contains {
		embeddedMigrationItem.removeLink(embedded)
		migrationMap[embeddedMigrationItem.NormalizedName] = embeddedMigrationItem
	}
}

func (s *Service) checkEmbeddedSourceEntry(
	migrationMap map[string]*MigrationItem,
	embedded string,
	item *MigrationItem,
) {
	embeddedMigrationItem, ok := migrationMap[embedded]
	if !ok ||
		(embeddedMigrationItem.CatalogInFile != nil &&
			arrayutil.Contains(embeddedMigrationItem.CatalogInFile.Embeds, item.NormalizedName)) ||
		(embeddedMigrationItem.CatalogInStore != nil &&
			arrayutil.Contains(embeddedMigrationItem.CatalogInStore.Embeds, item.NormalizedName)) {
		return
	}
	item.removeLink(embedded)
}

func (s *Service) checkEmbeddingOnRename(
	item *MigrationItem,
	migrationMap map[string]*MigrationItem,
) {
	for _, embedding := range item.CatalogInFile.Embeds {
		existing, ok := migrationMap[embedding]
		if !ok {
			continue
		}

		existing.CatalogInStore.Embeds = arrayutil.Delete(existing.CatalogInStore.Embeds, item.FromNormalizedName)
		existing.CatalogInStore.Links--
	}
}
