package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/arrayutil"
	"github.com/rilldata/rill/runtime/pkg/dag"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"

	// Load migrators
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/sql"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/yaml"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/metricsviews"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/models"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
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
	MigrationNoChange int = 0
	MigrationCreate   int = 1
	MigrationRename   int = 2
	MigrationUpdate   int = 3
	MigrationDelete   int = 4
)

type ReconcileConfig struct {
	DryRun       bool
	Strict       bool
	ChangedPaths []string
	ForcedPaths  []string
}

type ReconcileResult struct {
	AddedObjects   []*drivers.CatalogEntry
	UpdatedObjects []*drivers.CatalogEntry
	DroppedObjects []*drivers.CatalogEntry
	AffectedPaths  []string
	Errors         []*runtimev1.ReconcileError
}

func NewReconcileResult() *ReconcileResult {
	return &ReconcileResult{
		AddedObjects:   make([]*drivers.CatalogEntry, 0),
		UpdatedObjects: make([]*drivers.CatalogEntry, 0),
		DroppedObjects: make([]*drivers.CatalogEntry, 0),
		AffectedPaths:  make([]string, 0),
		Errors:         make([]*runtimev1.ReconcileError, 0),
	}
}

func (r *ReconcileResult) collectAffectedPaths() {
	pathDuplicates := make(map[string]bool)
	for _, added := range r.AddedObjects {
		r.AffectedPaths = append(r.AffectedPaths, added.Path)
		pathDuplicates[added.Path] = true
	}
	for _, updated := range r.UpdatedObjects {
		if pathDuplicates[updated.Path] {
			continue
		}
		r.AffectedPaths = append(r.AffectedPaths, updated.Path)
		pathDuplicates[updated.Path] = true
	}
	for _, deleted := range r.DroppedObjects {
		if pathDuplicates[deleted.Path] {
			continue
		}
		r.AffectedPaths = append(r.AffectedPaths, deleted.Path)
		pathDuplicates[deleted.Path] = true
	}
	for _, errored := range r.Errors {
		if pathDuplicates[errored.FilePath] {
			continue
		}
		r.AffectedPaths = append(r.AffectedPaths, errored.FilePath)
		pathDuplicates[errored.FilePath] = true
	}
}

type ArtifactError struct {
	Error error
	Path  string
}

// TODO: support loading existing projects

func (s *Service) Reconcile(ctx context.Context, conf ReconcileConfig) (*ReconcileResult, error) {
	result := NewReconcileResult()

	// collect repos and create migration items
	migrationMap, err := s.collectRepos(ctx, conf, result)
	if err != nil {
		return nil, err
	}

	// order the items to have parents before children
	migrations, err := s.collectMigrationItems(migrationMap, result)
	if err != nil {
		return nil, err
	}

	err = s.runMigrationItems(ctx, conf, migrations, result)
	if err != nil {
		return nil, err
	}

	if !conf.DryRun {
		// TODO: changes to the file will not be picked up if done while running migration
		s.LastMigration = time.Now()
	}
	result.collectAffectedPaths()
	return result, nil
}

// convert repo paths to MigrationItem

func (s *Service) collectRepos(ctx context.Context, conf ReconcileConfig, result *ReconcileResult) (map[string]*MigrationItem, error) {
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
			return nil, err
		}
	}

	forcedPathMap := make(map[string]bool)
	for _, forcedPath := range conf.ForcedPaths {
		forcedPathMap[forcedPath] = true
	}

	storeObjectsMap := make(map[string]*drivers.CatalogEntry)
	storeObjectsConsumed := make(map[string]bool)
	storeObjects := s.Catalog.FindEntries(ctx, s.InstID, drivers.ObjectTypeUnspecified)
	for _, storeObject := range storeObjects {
		storeObjectsMap[strings.ToLower(storeObject.Name)] = storeObject
	}

	migrationMap := make(map[string]*MigrationItem)
	deletions := make(map[string]*MigrationItem)
	additions := make(map[string]*MigrationItem)

	for _, repoPath := range repoPaths {
		item := s.getMigrationItem(ctx, repoPath, storeObjectsMap, forcedPathMap)
		if item == nil {
			continue
		}

		keepNew, errPath := s.isInvalidDuplicate(migrationMap, changedPathsHint, changedPathsMap, item)
		if errPath != "" {
			result.Errors = append(result.Errors, &runtimev1.ReconcileError{
				Code:     runtimev1.ReconcileError_CODE_UNSPECIFIED,
				Message:  "item with same name exists",
				FilePath: errPath,
			})
		}
		if !keepNew {
			continue
		}

		add := true
		switch item.Type {
		case MigrationCreate:
			// if item is created compare with deletions to look for renames
			found := false
			for _, deletion := range deletions {
				if migrator.IsEqual(ctx, item.CatalogInFile, deletion.CatalogInStore) {
					item.renameFrom(deletion)
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
					addition.renameFrom(item)
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

		if add {
			migrationMap[item.NormalizedName] = item
		}
		storeObjectsConsumed[item.NormalizedName] = true

		if !changedPathsHint {
			continue
		}
		// go through the children only if forced paths is false
		// children, err := s.dag.GetChildren(item.NormalizedName)
		// if err != nil {
		// 	return nil, err
		// }

		for _, child := range s.dag.GetChildren(item.NormalizedName) {
			childPath, ok := s.NameToPath[child]
			if !ok || (changedPathsHint && changedPathsMap[childPath]) {
				// if there is no entry for name to path or already in forced path then ignore the child
				continue
			}

			childItem := s.getMigrationItem(ctx, childPath, storeObjectsMap, forcedPathMap)
			if childItem == nil {
				continue
			}
			migrationMap[childItem.NormalizedName] = childItem
		}
	}

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
		found := false
		// find any additions that match and mark it as a MigrationRename
		for _, addition := range additions {
			if migrator.IsEqual(ctx, addition.CatalogInFile, storeObject) {
				addition.Type = MigrationRename
				addition.FromName = storeObject.Name
				addition.FromPath = storeObject.Path
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

	return migrationMap, nil
}

func (s *Service) getMigrationItem(
	ctx context.Context,
	repoPath string,
	storeObjectsMap map[string]*drivers.CatalogEntry,
	forcedPathMap map[string]bool,
) *MigrationItem {
	item := &MigrationItem{
		Type: MigrationNoChange,
		Path: repoPath,
	}

	forceChange := forcedPathMap[repoPath]

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

		item.NormalizedDependencies = migrator.GetDependencies(ctx, s.Olap, catalog)
		// convert dependencies to lower case
		for i, dep := range item.NormalizedDependencies {
			item.NormalizedDependencies[i] = strings.ToLower(dep)
		}
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
		return item
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

	return item
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
		if existingPath, ok := s.NameToPath[item.NormalizedName]; ok && existingPath != item.Path && !changedPathsMap[existingPath] {
			return false, item.Path
		}
	}

	return true, errPath
}

// collectMigrationItems collects all valid MigrationItem
// It will order the items based on DAG with parents coming before children.
func (s *Service) collectMigrationItems(
	migrationMap map[string]*MigrationItem,
	result *ReconcileResult,
) ([]*MigrationItem, error) {
	migrationItems := make([]*MigrationItem, 0)
	visited := make(map[string]int)
	update := make(map[string]bool)

	// temporary local dag for just the items to be migrated
	// this will also help in getting a dag for new items
	// TODO: is there a better way to do this?
	tempDag := dag.NewDAG()
	for name, migration := range migrationMap {
		_, err := tempDag.Add(name, migration.NormalizedDependencies)
		if err != nil {
			result.Errors = append(result.Errors, &runtimev1.ReconcileError{
				Code:     runtimev1.ReconcileError_CODE_SOURCE,
				Message:  err.Error(),
				FilePath: migration.Path,
			})
		}
	}

	for name, item := range migrationMap {
		if item.Type == MigrationNoChange {
			if update[name] {
				// items identified as to created/updated because a parent changed
				// but was initially marked no change
				if item.CatalogInStore == nil {
					item.Type = MigrationCreate
				} else {
					item.Type = MigrationUpdate
				}
			} else if _, ok := s.PathToName[item.Path]; ok {
				// this allows parents later in the order to re add children
				visited[name] = -1
				continue
			}
		}

		visited[name] = len(migrationItems)
		migrationItems = append(migrationItems, item)
		children := arrayutil.Dedupe(append(
			tempDag.GetChildren(name),
			s.dag.GetChildren(name)...,
		))
		if item.FromName != "" {
			children = append(children, arrayutil.Dedupe(append(
				tempDag.GetChildren(strings.ToLower(item.FromName)),
				s.dag.GetChildren(strings.ToLower(item.FromName))...,
			))...)
		}
		for _, child := range children {
			i, ok := visited[child]
			if !ok {
				// if not already visited, mark the child as needing update
				update[child] = true
				continue
			}

			var childItem *MigrationItem
			// if a child was already visited push to the end
			visited[child] = len(migrationItems)
			if i != -1 {
				childItem = migrationItems[i]
				// mark the original position as nil. this is cleaned up later
				migrationItems[i] = nil
			} else {
				childItem = migrationMap[child]
			}

			migrationItems = append(migrationItems, childItem)
			if childItem.Type == MigrationNoChange || childItem.Error != nil {
				// if the child has no change then mark it as update or create based on presence of catalog in store
				if childItem.CatalogInStore == nil {
					childItem.Type = MigrationCreate
				} else {
					childItem.Type = MigrationUpdate
				}
			}
		}
	}

	// cleanup any nil values that occurred by pushing child later in the order
	cleanedMigrationItems := make([]*MigrationItem, 0)
	for _, migration := range migrationItems {
		if migration == nil {
			continue
		}
		cleanedMigrationItems = append(cleanedMigrationItems, migration)
	}

	return cleanedMigrationItems, nil
}

// runMigrationItems runs various actions from MigrationItem based on MigrationItem.Type.
func (s *Service) runMigrationItems(
	ctx context.Context,
	conf ReconcileConfig,
	migrations []*MigrationItem,
	result *ReconcileResult,
) error {
	for _, item := range migrations {
		if item.Error != nil {
			result.Errors = append(result.Errors, item.Error)
		}

		var validationErrors []*runtimev1.ReconcileError

		if item.CatalogInFile != nil {
			validationErrors = migrator.Validate(ctx, s.Olap, item.CatalogInFile)
		}

		var err error
		failed := false
		if len(validationErrors) > 0 {
			// do not run migration if validation failed
			result.Errors = append(result.Errors, validationErrors...)
			failed = true
		} else if !conf.DryRun {
			if item.CatalogInStore != nil {
				// make sure store catalog has the correct name
				// could be different in cases like "rename with different case"
				item.CatalogInStore.Name = item.Name
			}
			// only run the actual migration if in dry run
			switch item.Type {
			case MigrationNoChange:
				if _, ok := s.PathToName[item.NormalizedName]; !ok {
					// this is perhaps an init. so populate cache data
					s.PathToName[item.Path] = item.NormalizedName
					s.NameToPath[item.NormalizedName] = item.Path
					_, err := s.dag.Add(item.NormalizedName, item.NormalizedDependencies)
					if err != nil {
						result.Errors = append(result.Errors, &runtimev1.ReconcileError{
							Code:     runtimev1.ReconcileError_CODE_SOURCE,
							Message:  err.Error(),
							FilePath: item.Path,
						})
					}
				}
			case MigrationCreate:
				err = s.createInStore(ctx, item, result)
				result.AddedObjects = append(result.AddedObjects, item.CatalogInFile)
			case MigrationRename:
				err = s.renameInStore(ctx, item, result)
				result.UpdatedObjects = append(result.UpdatedObjects, item.CatalogInFile)
			case MigrationUpdate:
				err = s.updateInStore(ctx, item, result)
				result.UpdatedObjects = append(result.UpdatedObjects, item.CatalogInFile)
			case MigrationDelete:
				err = s.deleteInStore(ctx, item)
				result.DroppedObjects = append(result.DroppedObjects, item.CatalogInStore)
			}
		}

		if err != nil {
			result.Errors = append(result.Errors, &runtimev1.ReconcileError{
				Code:     runtimev1.ReconcileError_CODE_OLAP,
				Message:  err.Error(),
				FilePath: item.Path,
			})
			failed = true
		}

		if failed && !conf.DryRun {
			// remove entity from catalog and OLAP if it failed validation or during migration
			err := s.Catalog.DeleteEntry(ctx, s.InstID, item.Name)
			if err != nil {
				// shouldn't ideally happen
				result.Errors = append(result.Errors, &runtimev1.ReconcileError{
					Code:     runtimev1.ReconcileError_CODE_OLAP,
					Message:  err.Error(),
					FilePath: item.Path,
				})
			}
			if item.CatalogInFile != nil {
				err := migrator.Delete(ctx, s.Olap, item.CatalogInFile)
				if err != nil {
					// shouldn't ideally happen
					result.Errors = append(result.Errors, &runtimev1.ReconcileError{
						Code:     runtimev1.ReconcileError_CODE_OLAP,
						Message:  err.Error(),
						FilePath: item.Path,
					})
				}
			}
			if conf.Strict {
				return err
			}
		}
	}

	return nil
}

// TODO: should we remove from dag if validation fails?
// TODO: store only valid metrics view

func (s *Service) createInStore(ctx context.Context, item *MigrationItem, result *ReconcileResult) error {
	s.NameToPath[item.NormalizedName] = item.Path
	s.PathToName[item.Path] = item.NormalizedName
	// add the item to DAG
	_, err := s.dag.Add(item.NormalizedName, item.NormalizedDependencies)
	if err != nil {
		result.Errors = append(result.Errors, &runtimev1.ReconcileError{
			Code:     runtimev1.ReconcileError_CODE_SOURCE,
			Message:  err.Error(),
			FilePath: item.Path,
		})
	}

	// create in olap
	err = s.wrapMigrator(item.CatalogInFile, func() error {
		return migrator.Create(ctx, s.Olap, s.Repo, item.CatalogInFile)
	})
	if err != nil {
		return err
	}

	// update the catalog object and create it in store
	catalog, err := s.updateCatalogObject(ctx, item)
	if err != nil {
		return err
	}
	_, found := s.Catalog.FindEntry(ctx, s.InstID, item.Name)
	// create or updated
	if found {
		return s.Catalog.UpdateEntry(ctx, s.InstID, catalog)
	}
	return s.Catalog.CreateEntry(ctx, s.InstID, catalog)
}

func (s *Service) renameInStore(ctx context.Context, item *MigrationItem, result *ReconcileResult) error {
	fromLowerName := strings.ToLower(item.FromName)
	delete(s.NameToPath, fromLowerName)

	s.NameToPath[item.NormalizedName] = item.Path
	delete(s.PathToName, item.FromPath)

	s.PathToName[item.Path] = item.NormalizedName

	// delete old item and add new item to dag
	s.dag.Delete(fromLowerName)
	_, err := s.dag.Add(item.NormalizedName, item.NormalizedDependencies)
	if err != nil {
		result.Errors = append(result.Errors, &runtimev1.ReconcileError{
			Code:     runtimev1.ReconcileError_CODE_SOURCE,
			Message:  err.Error(),
			FilePath: item.Path,
		})
	}

	// rename the item in olap
	err = migrator.Rename(ctx, s.Olap, item.FromName, item.CatalogInFile)
	if err != nil {
		return err
	}

	// delete the old catalog object
	// TODO: do we need a rename here?
	err = s.Catalog.DeleteEntry(ctx, s.InstID, item.FromName)
	if err != nil {
		return err
	}
	// update the catalog object and create it in store
	catalog, err := s.updateCatalogObject(ctx, item)
	if err != nil {
		return err
	}
	return s.Catalog.CreateEntry(ctx, s.InstID, catalog)
}

func (s *Service) updateInStore(ctx context.Context, item *MigrationItem, result *ReconcileResult) error {
	s.NameToPath[item.NormalizedName] = item.Path
	s.PathToName[item.Path] = item.NormalizedName
	// add the item to DAG with new dependencies
	_, err := s.dag.Add(item.NormalizedName, item.NormalizedDependencies)
	if err != nil {
		result.Errors = append(result.Errors, &runtimev1.ReconcileError{
			Code:     runtimev1.ReconcileError_CODE_SOURCE,
			Message:  err.Error(),
			FilePath: item.Path,
		})
	}

	// update in olap
	err = s.wrapMigrator(item.CatalogInFile, func() error {
		return migrator.Update(ctx, s.Olap, s.Repo, item.CatalogInFile)
	})
	if err != nil {
		return err
	}
	// update the catalog object and update it in store
	catalog, err := s.updateCatalogObject(ctx, item)
	if err != nil {
		return err
	}
	return s.Catalog.UpdateEntry(ctx, s.InstID, catalog)
}

func (s *Service) deleteInStore(ctx context.Context, item *MigrationItem) error {
	delete(s.NameToPath, item.NormalizedName)
	delete(s.PathToName, item.FromPath)

	// delete item from dag
	s.dag.Delete(item.NormalizedName)
	// delete item from olap
	err := migrator.Delete(ctx, s.Olap, item.CatalogInStore)
	if err != nil {
		return err
	}

	// delete from catalog store
	return s.Catalog.DeleteEntry(ctx, s.InstID, item.Name)
}

func (s *Service) updateCatalogObject(ctx context.Context, item *MigrationItem) (*drivers.CatalogEntry, error) {
	// get artifact stats
	repoStat, err := s.Repo.Stat(ctx, s.InstID, item.Path)
	if err != nil {
		return nil, err
	}

	// convert protobuf to database object
	catalogEntry := item.CatalogInFile
	// NOTE: Previously there was a copy here when using the API types. This might have to reverted.

	// set the UpdatedOn as LastUpdated from the artifact file
	// this will allow to not reprocess unchanged files
	catalogEntry.UpdatedOn = repoStat.LastUpdated
	catalogEntry.RefreshedOn = time.Now()

	err = migrator.SetSchema(ctx, s.Olap, catalogEntry)
	if err != nil {
		return nil, err
	}

	return catalogEntry, nil
}

// wrapMigrator is a temporary solution to log source related messages.
func (s *Service) wrapMigrator(catalogEntry *drivers.CatalogEntry, run func() error) error {
	if catalogEntry.Type == drivers.ObjectTypeSource {
		s.logger.Info(fmt.Sprintf(
			"Ingesting source %q from %q",
			catalogEntry.Name, catalogEntry.GetSource().Properties.Fields["path"].GetStringValue(),
		))
	}
	err := run()
	if catalogEntry.Type == drivers.ObjectTypeSource {
		if err != nil {
			s.logger.Error(fmt.Sprintf("Ingestion failed for %q : %s", catalogEntry.Name, err.Error()))
		} else {
			s.logger.Info(fmt.Sprintf("Finished ingesting %q", catalogEntry.Name))
		}
	}
	return err
}
