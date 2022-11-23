package catalog

import (
	"context"
	"time"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/arrayutil"
	"github.com/rilldata/rill/runtime/pkg/dag"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/sql"
	_ "github.com/rilldata/rill/runtime/services/catalog/artifacts/yaml"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/metrics_views"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/models"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MigrationItem struct {
	Name           string
	Path           string
	CatalogInFile  *api.CatalogObject
	CatalogInStore *api.CatalogObject
	Type           int
	FromName       string
	FromPath       string
	Dependencies   []string
	Error          *api.MigrationError
}

func (i *MigrationItem) renameFrom(from *MigrationItem) {
	i.Type = MigrationRename
	i.FromName = from.Name
	i.FromPath = from.Path
}

const (
	MigrationNoChange int = 0
	MigrationCreate       = 1
	MigrationRename       = 2
	MigrationUpdate       = 3
	MigrationDelete       = 4
)

type MigrationConfig struct {
	DryRun       bool
	Strict       bool
	ChangedPaths []string
	ForcedPaths  []string
}

type MigrationResult struct {
	AddedObjects   []*api.CatalogObject
	UpdatedObjects []*api.CatalogObject
	DroppedObjects []*api.CatalogObject
	AffectedPaths  []string
	Errors         []*api.MigrationError
}

func NewMigrationResult() *MigrationResult {
	return &MigrationResult{
		AddedObjects:   make([]*api.CatalogObject, 0),
		UpdatedObjects: make([]*api.CatalogObject, 0),
		DroppedObjects: make([]*api.CatalogObject, 0),
		AffectedPaths:  make([]string, 0),
		Errors:         make([]*api.MigrationError, 0),
	}
}

func (r *MigrationResult) collectAffectedPaths() {
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

func (s *Service) Migrate(
	ctx context.Context,
	conf MigrationConfig,
) (*MigrationResult, error) {
	result := NewMigrationResult()

	// collect repos and create migration items
	migrationMap, err := s.collectRepos(ctx, conf, result)
	if err != nil {
		return nil, err
	}

	// order the items to have parents before children
	migrations := s.collectMigrationItems(migrationMap)

	if conf.DryRun {
		return result, nil
	}

	err = s.runMigrationItems(ctx, conf, migrations, result)
	if err != nil {
		return nil, err
	}

	// TODO: changes to the file will not be picked up if done while running migration
	s.LastMigration = time.Now()
	result.collectAffectedPaths()
	return result, nil
}

// convert repo paths to MigrationItem

func (s *Service) collectRepos(ctx context.Context, conf MigrationConfig, result *MigrationResult) (map[string]*MigrationItem, error) {
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
		repoPaths, err = s.Repo.ListRecursive(ctx, s.RepoId, "{sources,models,dashboards}/*.{sql,yaml,yml}")
		if err != nil {
			return nil, err
		}
	}

	forcedPathMap := make(map[string]bool)
	for _, forcedPath := range conf.ForcedPaths {
		forcedPathMap[forcedPath] = true
	}

	storeObjectsMap := make(map[string]*drivers.CatalogObject)
	storeObjectsConsumed := make(map[string]bool)
	storeObjects := s.Catalog.FindObjects(ctx, s.InstId, drivers.CatalogObjectTypeUnspecified)
	for _, storeObject := range storeObjects {
		storeObjectsMap[storeObject.Name] = storeObject
	}

	migrationMap := make(map[string]*MigrationItem)
	deletions := make(map[string]*MigrationItem)
	additions := make(map[string]*MigrationItem)

	for _, repoPath := range repoPaths {
		item := s.getMigrationItem(ctx, repoPath, storeObjectsMap, forcedPathMap)
		if item == nil {
			continue
		}

		existing, ok := migrationMap[item.Name]
		if ok {
			var errPath string
			// if existing item was deleted
			if existing.Type == MigrationDelete ||
				// or if the existing has error whereas new one doest
				(item.Error != nil && existing.Error != nil) ||
				// or if the existing file was updated after new (this makes it so that the old one will be retained)
				(item.Error == nil && item.CatalogInFile.UpdatedOn != nil && existing.CatalogInFile.UpdatedOn != nil &&
					existing.CatalogInFile.UpdatedOn.AsTime().After(item.CatalogInFile.UpdatedOn.AsTime())) {
				// replace the existing with new
				migrationMap[item.Name] = item
				errPath = existing.Path
			} else {
				errPath = item.Path
			}
			result.Errors = append(result.Errors, &api.MigrationError{
				Code:     api.MigrationError_CODE_UNSPECIFIED,
				Message:  "item with same name exists",
				FilePath: errPath,
			})
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
					delete(deletions, deletion.Name)
					delete(migrationMap, deletion.Name)
					found = true
					break
				}
			}
			if !found {
				additions[item.Name] = item
			}

		case MigrationDelete:
			found := false
			// if item is deleted compare with additions to look for renames
			for _, addition := range additions {
				if migrator.IsEqual(ctx, addition.CatalogInFile, item.CatalogInStore) {
					addition.renameFrom(item)
					delete(additions, addition.Name)
					add = false
					found = true
					break
				}
			}
			if !found {
				deletions[item.Name] = item
			}
		}

		if add {
			migrationMap[item.Name] = item
		}
		storeObjectsConsumed[item.Name] = true

		if !changedPathsHint {
			continue
		}
		// go through the children only of forced paths is false
		children := s.dag.GetChildren(item.Name)
		for _, child := range children {
			childPath, ok := s.NameToPath[child]
			if !ok || (changedPathsHint && changedPathsMap[childPath]) {
				// if there is no entry for name to path or already in forced path then ignore the child
				continue
			}

			childItem := s.getMigrationItem(ctx, childPath, storeObjectsMap, forcedPathMap)
			if childItem == nil {
				continue
			}
			migrationMap[childItem.Name] = childItem
		}
	}

	for _, storeObject := range storeObjectsMap {
		// ignore consumed store objects
		if storeObjectsConsumed[storeObject.Name] ||
			// ignore tables and unspecified objects
			storeObject.Type == drivers.CatalogObjectTypeTable || storeObject.Type == drivers.CatalogObjectTypeUnspecified {
			continue
		}
		apiCatalog, err := catalogObjectToPB(storeObject)
		if err != nil {
			continue
		}
		// if repo paths were forced and the catalog was not in the paths then ignore
		if _, ok := changedPathsMap[apiCatalog.Path]; changedPathsHint && !ok {
			continue
		}
		found := false
		// find any additions that match and mark it as a MigrationRename
		for _, addition := range additions {
			if migrator.IsEqual(ctx, addition.CatalogInFile, apiCatalog) {
				addition.Type = MigrationRename
				addition.FromName = apiCatalog.Name
				addition.FromPath = apiCatalog.Path
				delete(additions, addition.Name)
				found = true
				break
			}
		}
		// if no matching item is found, add as a MigrationDelete
		if !found {
			migrationMap[apiCatalog.Name] = &MigrationItem{
				Name:           apiCatalog.Name,
				Type:           MigrationDelete,
				Path:           apiCatalog.Path,
				CatalogInStore: apiCatalog,
			}
		}
	}

	return migrationMap, nil
}

func (s *Service) getMigrationItem(
	ctx context.Context,
	repoPath string,
	storeObjectsMap map[string]*drivers.CatalogObject,
	forcedPathMap map[string]bool,
) *MigrationItem {
	item := &MigrationItem{
		Type: MigrationNoChange,
		Path: repoPath,
	}

	catalog, err := artifacts.Read(ctx, s.Repo, s.RepoId, repoPath)
	if err != nil {
		if err != artifacts.FileReadError {
			item.Error = &api.MigrationError{
				Code:     api.MigrationError_CODE_SYNTAX,
				Message:  err.Error(),
				FilePath: repoPath,
			}
		}
		if _, ok := s.PathToName[repoPath]; !ok {
			return nil
		}

		item.Name = s.PathToName[repoPath]
		item.Type = MigrationDelete
	} else {
		item.Name = catalog.Name
		item.CatalogInFile = catalog

		item.Dependencies = migrator.GetDependencies(ctx, s.Olap, catalog)
		repoStat, _ := s.Repo.Stat(ctx, s.RepoId, repoPath)
		item.CatalogInFile.UpdatedOn = timestamppb.New(repoStat.LastUpdated)
		if repoStat.LastUpdated.After(s.LastMigration) {
			// assume creation until we see a catalog object
			item.Type = MigrationCreate
		}
	}

	catalogInStore, ok := storeObjectsMap[item.Name]
	if !ok {
		return item
	}
	apiCatalog, err := catalogObjectToPB(catalogInStore)
	if err != nil {
		return item
	}
	item.CatalogInStore = apiCatalog

	switch item.Type {
	case MigrationCreate:
		if migrator.IsEqual(ctx, item.CatalogInFile, item.CatalogInStore) && !forcedPathMap[repoPath] {
			// if the actual content has not changed, mark as MigrationNoChange
			item.Type = MigrationNoChange
		} else {
			// else mark as MigrationUpdate
			item.Type = MigrationUpdate
		}

	case MigrationNoChange:
		if forcedPathMap[repoPath] {
			item.Type = MigrationUpdate
			break
		}

		// if item doesnt exist in olap, mark as create
		ok, _ := migrator.ExistsInOlap(ctx, s.Olap, item.CatalogInFile)
		if !ok {
			item.Type = MigrationCreate
		}
	}

	return item
}

// collectMigrationItems collects all valid MigrationItem
// It will order the items based on DAG with parents coming before children
func (s *Service) collectMigrationItems(
	migrationMap map[string]*MigrationItem,
) []*MigrationItem {
	migrationItems := make([]*MigrationItem, 0)
	visited := make(map[string]int)
	update := make(map[string]bool)

	// temporary local dag for just the items to be migrated
	// this will also help in getting a dag for new items
	// TODO: is there a better way to do this?
	tempDag := dag.NewDAG()
	for name, migration := range migrationMap {
		tempDag.Add(name, migration.Dependencies)
	}

	for name, migration := range migrationMap {
		if migration.Type == MigrationNoChange {
			if update[name] {
				// items identified as to created/updated because a parent changed
				// but was initially marked no change
				if migration.CatalogInStore == nil {
					migration.Type = MigrationCreate
				} else {
					migration.Type = MigrationUpdate
				}
			} else if _, ok := s.PathToName[migration.Path]; ok {
				// this allows parents later in the order to re add children
				visited[name] = -1
				continue
			}
		}

		visited[name] = len(migrationItems)
		migrationItems = append(migrationItems, migration)

		// get all the children and make sure they are not present before the parent in the order
		children := arrayutil.Dedupe(append(
			tempDag.GetChildren(name),
			s.dag.GetChildren(name)...,
		))
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

	return cleanedMigrationItems
}

// runMigrationItems runs various actions from MigrationItem based on MigrationItem.Type
func (s *Service) runMigrationItems(
	ctx context.Context,
	conf MigrationConfig,
	migrations []*MigrationItem,
	result *MigrationResult,
) error {
	for _, item := range migrations {
		var err error

		if item.CatalogInFile != nil {
			err = migrator.Validate(ctx, s.Olap, item.CatalogInFile)
		}

		if err == nil {
			switch item.Type {
			case MigrationNoChange:
				if _, ok := s.PathToName[item.Path]; !ok {
					// this is perhaps an init. so populate cache data
					s.PathToName[item.Path] = item.Name
					s.NameToPath[item.Name] = item.Path
					s.dag.Add(item.Name, item.Dependencies)
				}
			case MigrationCreate:
				err = s.createInStore(ctx, item)
				result.AddedObjects = append(result.AddedObjects, item.CatalogInFile)
			case MigrationRename:
				err = s.renameInStore(ctx, item)
				result.UpdatedObjects = append(result.UpdatedObjects, item.CatalogInFile)
			case MigrationUpdate:
				err = s.updateInStore(ctx, item)
				result.UpdatedObjects = append(result.UpdatedObjects, item.CatalogInFile)
			case MigrationDelete:
				err = s.deleteInStore(ctx, item)
				result.DroppedObjects = append(result.DroppedObjects, item.CatalogInStore)
			}
		}

		if err != nil {
			result.Errors = append(result.Errors, &api.MigrationError{
				Code:     api.MigrationError_CODE_OLAP,
				Message:  err.Error(),
				FilePath: item.Path,
			})
			err := s.Catalog.DeleteObject(ctx, s.InstId, item.Name)
			if err != nil {
				// shouldn't ideally happen
				result.Errors = append(result.Errors, &api.MigrationError{
					Code:     api.MigrationError_CODE_OLAP,
					Message:  err.Error(),
					FilePath: item.Path,
				})
			}
			if item.CatalogInFile != nil {
				err := migrator.Delete(ctx, s.Olap, item.CatalogInFile)
				if err != nil {
					// shouldn't ideally happen
					result.Errors = append(result.Errors, &api.MigrationError{
						Code:     api.MigrationError_CODE_OLAP,
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

func (s *Service) createInStore(ctx context.Context, item *MigrationItem) error {
	s.NameToPath[item.Name] = item.Path
	s.PathToName[item.Path] = item.Name
	// add the item to DAG
	s.dag.Add(item.Name, item.Dependencies)

	// create in olap
	err := migrator.Create(ctx, s.Olap, s.Repo, item.CatalogInFile)
	if err != nil {
		return err
	}

	// update the catalog object and create it in store
	catalog, err := s.updateCatalogObject(ctx, item)
	if err != nil {
		return err
	}
	_, found := s.Catalog.FindObject(ctx, s.InstId, item.Name)
	// create or updated
	if found {
		return s.Catalog.UpdateObject(ctx, s.InstId, catalog)
	} else {
		return s.Catalog.CreateObject(ctx, s.InstId, catalog)
	}
}

func (s *Service) renameInStore(ctx context.Context, item *MigrationItem) error {
	if _, ok := s.NameToPath[item.FromName]; ok {
		delete(s.NameToPath, item.FromName)
	}
	s.NameToPath[item.Name] = item.Path
	if _, ok := s.PathToName[item.FromPath]; ok {
		delete(s.PathToName, item.FromPath)
	}
	s.PathToName[item.Path] = item.Name

	// delete old item and add new item to dag
	s.dag.Delete(item.FromName)
	s.dag.Add(item.Name, item.Dependencies)

	// rename the item in olap
	err := migrator.Rename(ctx, s.Olap, item.FromName, item.CatalogInFile)
	if err != nil {
		return err
	}

	// delete the old catalog object
	// TODO: do we need a rename here?
	err = s.Catalog.DeleteObject(ctx, s.InstId, item.FromName)
	// update the catalog object and create it in store
	catalog, err := s.updateCatalogObject(ctx, item)
	if err != nil {
		return err
	}
	return s.Catalog.CreateObject(ctx, s.InstId, catalog)
}

func (s *Service) updateInStore(ctx context.Context, item *MigrationItem) error {
	s.NameToPath[item.Name] = item.Path
	s.PathToName[item.Path] = item.Name
	// add the item to DAG with new dependencies
	s.dag.Add(item.Name, item.Dependencies)

	// update in olap
	err := migrator.Update(ctx, s.Olap, s.Repo, item.CatalogInFile)
	if err != nil {
		return err
	}
	// update the catalog object and update it in store
	catalog, err := s.updateCatalogObject(ctx, item)
	if err != nil {
		return err
	}
	return s.Catalog.UpdateObject(ctx, s.InstId, catalog)
}

func (s *Service) deleteInStore(ctx context.Context, item *MigrationItem) error {
	if _, ok := s.NameToPath[item.Name]; ok {
		delete(s.NameToPath, item.Name)
	}
	if _, ok := s.PathToName[item.FromPath]; ok {
		delete(s.PathToName, item.FromPath)
	}

	// delete item from dag
	s.dag.Delete(item.Name)
	// delete item from olap
	err := migrator.Delete(ctx, s.Olap, item.CatalogInStore)
	if err != nil {
		return err
	}

	// delete from catalog store
	return s.Catalog.DeleteObject(ctx, s.InstId, item.Name)
}

func (s *Service) updateCatalogObject(ctx context.Context, item *MigrationItem) (*drivers.CatalogObject, error) {
	// get artifact stats
	repoStat, err := s.Repo.Stat(ctx, s.RepoId, item.Path)
	if err != nil {
		return nil, err
	}

	// convert protobuf to database object
	catalog, err := pbToCatalogObject(item.CatalogInFile)
	if err != nil {
		return nil, err
	}

	// set the UpdatedOn as LastUpdated from the artifact file
	// this will allow to not reprocess unchanged files
	catalog.UpdatedOn = repoStat.LastUpdated
	catalog.RefreshedOn = time.Now()

	catalog.Schema, err = migrator.GetSchema(ctx, s.Olap, item.CatalogInFile)
	if err != nil {
		return nil, err
	}

	return catalog, nil
}
