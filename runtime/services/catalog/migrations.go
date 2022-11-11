package catalog

import (
	"context"
	"time"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/dag"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/metrics_views"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/models"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
)

type MigrationItem struct {
	Name           string
	Path           string
	CatalogInFile  *api.CatalogObject
	CatalogInStore *api.CatalogObject
	Type           int
	FromName       string
	Dependencies   []string
	Error          *ArtifactError
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
}

type MigrationResult struct {
	AddedObjects   []*api.CatalogObject
	UpdatedObjects []*api.CatalogObject
	DroppedObjects []*api.CatalogObject
	ArtifactErrors []ArtifactError
}

type ArtifactError struct {
	Error error
	Path  string
}

func (s *Service) Migrate(
	ctx context.Context,
	conf MigrationConfig,
) (MigrationResult, error) {
	result := MigrationResult{
		AddedObjects:   make([]*api.CatalogObject, 0),
		UpdatedObjects: make([]*api.CatalogObject, 0),
		DroppedObjects: make([]*api.CatalogObject, 0),
		ArtifactErrors: make([]ArtifactError, 0),
	}

	// get the initial name to migration item map
	migrationMap := make(map[string]*MigrationItem)
	var err error
	if len(conf.ChangedPaths) > 0 {
		migrationMap, err = s.collectSelectedRepos(ctx, conf.ChangedPaths)
	} else {
		migrationMap, err = s.collectAllRepos(ctx)
	}
	if err != nil {
		return result, err
	}

	// map the items to catalog objects
	s.collectCatalogObjects(ctx, conf, migrationMap)
	// order the items to have parents before children
	migrations := s.collectMigrationItems(migrationMap)

	if conf.DryRun {
		return result, nil
	}

	err = s.runMigrationItems(ctx, conf, migrations, &result)
	if err != nil {
		return result, err
	}

	// TODO: changes to the file will not be picked up if done while running migration
	s.LastMigration = time.Now()

	return result, nil
}

// convert repo paths to MigrationItem

func (s *Service) collectAllRepos(ctx context.Context) (map[string]*MigrationItem, error) {
	// TODO: if the repo folder is source controlled we should leverage it to find changes
	// TODO: ListRecursive needs some kind of cache or optimisation
	repoPaths, err := s.Repo.ListRecursive(ctx, s.RepoId)
	if err != nil {
		return nil, err
	}
	migrationMap := make(map[string]*MigrationItem)

	for _, repoPath := range repoPaths {
		item := s.getMigrationItem(ctx, repoPath)
		if item == nil {
			continue
		}
		migrationMap[item.Name] = item
	}

	return migrationMap, nil
}

func (s *Service) collectSelectedRepos(ctx context.Context, changedPaths []string) (map[string]*MigrationItem, error) {
	migrationMap := make(map[string]*MigrationItem)
	childRepos := make(map[string]bool)

	for _, repoPath := range changedPaths {
		// delete from childRepos if exists. this might be added by a parent processed before
		if _, ok := childRepos[repoPath]; ok {
			delete(childRepos, repoPath)
		}
		item := s.getMigrationItem(ctx, repoPath)
		if item == nil {
			continue
		}
		migrationMap[item.Name] = item

		// check children since the ChangedPaths might not have all the paths
		children := s.dag.GetChildren(item.Name)
		for _, child := range children {
			// if it was already visited then we dont need to add
			if _, ok := migrationMap[child]; ok {
				// force update if marked as NoChange
				if migrationMap[child].Type == MigrationNoChange {
					migrationMap[child].Type = MigrationUpdate
				}
				continue
			}
			childRepos[child] = true
		}
	}

	// go through all child repos and read them
	for child := range childRepos {
		path, ok := s.NameToPath[child]
		if !ok {
			// TODO: handle unknown child
			continue
		}

		item := s.getMigrationItem(ctx, path)
		if item == nil {
			continue
		}
		migrationMap[item.Name] = item
	}

	return migrationMap, nil
}

func (s *Service) getMigrationItem(ctx context.Context, repoPath string) *MigrationItem {
	catalog, err := artifacts.Read(ctx, s.Repo, s.RepoId, repoPath)
	if err != nil {
		// TODO: how do we get the name of the item?
		return nil
	}
	item := &MigrationItem{
		Name:          catalog.Name,
		Type:          MigrationNoChange,
		CatalogInFile: catalog,
		Path:          repoPath,
	}

	item.Dependencies = migrator.GetDependencies(ctx, s.Olap, catalog)
	err = migrator.Validate(ctx, s.Olap, catalog)
	if err != nil {
		item.Error = &ArtifactError{
			Error: err,
			Path:  repoPath,
		}
		return item
	}

	repoStat, _ := s.Repo.Stat(ctx, s.RepoId, repoPath)
	if repoStat.LastUpdated.After(s.LastMigration) {
		// assume creation until we see a catalog object
		item.Type = MigrationCreate
	}

	return item
}

// collectCatalogObjects maps catalog objects to existing MigrationItem
func (s *Service) collectCatalogObjects(
	ctx context.Context,
	conf MigrationConfig,
	migrationMap map[string]*MigrationItem,
) {
	catalogObjs := s.Catalog.FindObjects(ctx, s.InstId, drivers.CatalogObjectTypeUnspecified)
	deletions := make([]*MigrationItem, 0)

	for _, catalogObj := range catalogObjs {
		apiCatalog, err := catalogObjectToPB(catalogObj)
		if err != nil {
			// TODO
			continue
		}

		item, ok := migrationMap[apiCatalog.Name]
		if !ok {
			if len(conf.ChangedPaths) > 0 {
				// we wont have a match when specific paths are chosen
				continue
			}
			deletions = append(deletions, &MigrationItem{
				Name:          apiCatalog.Name,
				CatalogInFile: apiCatalog,
				Type:          MigrationDelete,
			})
		} else {
			// update store catalog
			item.CatalogInStore = apiCatalog
			if item.Type == MigrationCreate {
				item.Type = MigrationUpdate
				// if the actual content is same mark as no change
				if migrator.IsEqual(ctx, item.CatalogInFile, item.CatalogInStore) {
					item.Type = MigrationNoChange
				}
			} else if item.Error != nil {
				item.Type = MigrationDelete
			}
		}
	}

	for _, deletion := range deletions {
		renamed := false
		for _, item := range migrationMap {
			// only consider items that are equal and are marked as create
			if migrator.IsEqual(ctx, deletion.CatalogInFile, item.CatalogInFile) && item.Type == MigrationCreate {
				item.CatalogInStore.Type = MigrationRename
				item.FromName = deletion.CatalogInFile.Name
				renamed = true
				break
			}
		}

		// if not renamed then add as a deletion
		if !renamed {
			migrationMap[deletion.Name] = deletion
		}
	}
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
			} else {
				// this allows parents later in the order to re add children
				visited[name] = -1
				continue
			}
		}

		visited[name] = len(migrationItems)
		migrationItems = append(migrationItems, migration)

		// get all the children and make sure they are not before the parent in the order
		children := tempDag.GetChildren(name)
		for _, child := range children {
			i, ok := visited[child]
			if !ok {
				// if not already visited, mark the child as needing update
				update[child] = true
				continue
			}
			// if a child was already visited push to the end
			visited[child] = len(migrationItems)
			var childItem *MigrationItem
			if i != -1 {
				childItem = migrationItems[i]
				// mark the original position as nil. this is cleaned up later
				migrationItems[i] = nil
			} else {
				childItem = migrationMap[child]
			}
			migrationItems = append(migrationItems, childItem)
			if childItem.Type == MigrationNoChange {
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

		if item.CatalogInFile != nil && item.CatalogInFile.Type == api.CatalogObject_TYPE_METRICS_VIEW {
			err = migrator.Validate(ctx, s.Olap, item.CatalogInFile)
		}

		if err == nil {
			switch item.Type {
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
				result.DroppedObjects = append(result.DroppedObjects, item.CatalogInFile)
			}
		}

		if err != nil {
			result.ArtifactErrors = append(result.ArtifactErrors, ArtifactError{
				Error: err,
				Path:  item.Path,
			})
			err := s.Catalog.DeleteObject(ctx, s.InstId, item.Name)
			if err != nil {
				// shouldn't ideally happen
				result.ArtifactErrors = append(result.ArtifactErrors, ArtifactError{
					Error: err,
					Path:  item.Path,
				})
			}
			if item.CatalogInFile != nil {
				err := migrator.Delete(ctx, s.Olap, item.CatalogInFile)
				if err != nil {
					// shouldn't ideally happen
					result.ArtifactErrors = append(result.ArtifactErrors, ArtifactError{
						Error: err,
						Path:  item.Path,
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
	// add the item to DAG
	s.dag.Add(item.Name, item.Dependencies)

	// create in olap
	err := migrator.Create(ctx, s.Olap, item.CatalogInFile)
	if err != nil {
		return err
	}

	// update the catalog object and create it in store
	catalog, err := s.updateCatalogObject(ctx, item)
	if err != nil {
		return err
	}
	return s.Catalog.CreateObject(ctx, s.InstId, catalog)
}

func (s *Service) renameInStore(ctx context.Context, item *MigrationItem) error {
	if _, ok := s.NameToPath[item.FromName]; ok {
		delete(s.NameToPath, item.FromName)
	}
	s.NameToPath[item.Name] = item.Path

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
	// add the item to DAG with new dependencies
	s.dag.Add(item.Name, item.Dependencies)

	// update in olap
	err := migrator.Update(ctx, s.Olap, item.CatalogInFile)
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

	// delete item from dag
	s.dag.Delete(item.Name)
	// delete item from olap
	err := migrator.Delete(ctx, s.Olap, item.CatalogInFile)
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
	return catalog, nil
}
