package catalog

import (
	"context"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
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
	reconcileErrors := make([]*runtimev1.ReconcileError, 0)
	deletions := make(map[string]*MigrationItem)
	additions := make(map[string]*MigrationItem)

	for _, repoPath := range repoPaths {
		var items []*MigrationItem
		if embedded, ok := storeObjectsPathMap[repoPath]; ok && embedded.Embedded {
			// embedded items need to be created differently
			items = []*MigrationItem{s.newEmbeddedMigrationItem(embedded, MigrationUpdate)}
		} else {
			items = s.getMigrationItems(ctx, repoPath, storeObjectsMap, forcedPathMap)
		}
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

			if s.lookForRenames(ctx, item, migrationMap, additions, deletions) {
				migrationMap[item.NormalizedName] = item
			}
			storeObjectsConsumed[item.NormalizedName] = true

			if !changedPathsHint {
				continue
			}

			if item.Type == MigrationNoChange && item.CatalogInStore == nil && changedPathsMap[repoPath] {
				// new file added adhoc
				item.Type = MigrationCreate
			}

			// go through the children only if forced paths is false
			children := s.dag.GetDeepChildren(item.NormalizedName)
			for _, child := range children {
				childPath, ok := s.NameToPath[child]
				if !ok || (changedPathsHint && changedPathsMap[childPath]) {
					// if there is no entry for name to path or already in forced path then ignore the child
					continue
				}

				childItems := s.getMigrationItems(ctx, childPath, storeObjectsMap, forcedPathMap)
				for _, childItem := range childItems {
					migrationMap[childItem.NormalizedName] = childItem
				}
			}
		}
	}

	for _, storeEntry := range storeObjectsMap {
		normalisedStoreName := normalizeName(storeEntry.Name)
		// ignore consumed store objects
		if storeObjectsConsumed[normalisedStoreName] ||
			// ignore tables and unspecified objects
			storeEntry.Type == drivers.ObjectTypeTable || storeEntry.Type == drivers.ObjectTypeUnspecified {
			continue
		}
		// ignore embedded sources
		if storeEntry.Embedded {
			if _, added := migrationMap[normalisedStoreName]; !added && !s.hasMigrated {
				// only add if it was the 1st migration and has not already been added
				migrationMap[normalisedStoreName] = s.newEmbeddedMigrationItem(storeEntry, MigrationNoChange)
			}
			continue
		}
		// if repo paths were forced and the catalog was not in the paths then ignore
		if _, ok := changedPathsMap[storeEntry.Path]; changedPathsHint && !ok {
			continue
		}
		found := false
		// find any additions that match and mark it as a MigrationRename
		for _, addition := range additions {
			if migrator.IsEqual(ctx, addition.CatalogInFile, storeEntry) {
				addition.renameFrom(storeEntry.Name, storeEntry.Path)
				delete(additions, addition.NormalizedName)
				found = true
				break
			}
		}
		// if no matching item is found, add as a MigrationDelete
		if !found {
			migrationMap[normalisedStoreName] = s.newDeleteMigrationItem(storeEntry)
			parents := s.dag.GetParents(normalisedStoreName)
			for _, parent := range parents {
				_, migrating := migrationMap[parent]
				if migrating {
					continue
				}
				parentEntry, ok := storeObjectsMap[parent]
				if !ok || !parentEntry.Embedded {
					// only add embedded entries for now
					continue
				}
				migrationMap[parent] = s.newEmbeddedMigrationItem(parentEntry, MigrationReportUpdate)
			}
		}
	}

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
		if item.NewCatalog != nil && item.NewCatalog.Embedded {
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
