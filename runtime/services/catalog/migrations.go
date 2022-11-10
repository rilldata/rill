package catalog

import (
	"context"
	"time"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/models"
	_ "github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
)

type MigrationItem struct {
	CatalogInFile  *api.CatalogObject
	CatalogInStore *api.CatalogObject
	Path           string
	Type           int
	FromName       string
}

const (
	MigrationNoChange int = 0
	MigrationCreate       = 1
	MigrationRename       = 2
	MigrationUpdate       = 3
	MigrationDelete       = 4
)

type MigrationConfig struct {
	DryRun     bool
	BestEffort bool

	Renamed    string
	RenameFrom string
	ReIngest   string
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

	migrationMap, err := s.collectRepos(ctx, conf)
	if err != nil {
		return result, err
	}
	migrations := s.collectMigrationItems(ctx, conf, migrationMap)

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

func (s *Service) collectRepos(ctx context.Context, conf MigrationConfig) (map[string]*MigrationItem, error) {
	// TODO: if the repo folder is source controlled we should leverage it to find changes
	// TODO: ListRecursive needs some kind of cache or optimisation
	repoPaths, err := s.Repo.ListRecursive(ctx, s.RepoId)
	if err != nil {
		return nil, err
	}
	migrationMap := make(map[string]*MigrationItem)

	for _, repoPath := range repoPaths {
		catalog, err := artifacts.Read(ctx, s.Repo, s.RepoId, repoPath)
		if err != nil {
			// TODO
			continue
		}
		item := &MigrationItem{
			CatalogInFile: catalog,
			Path:          repoPath,
		}

		if conf.Renamed == catalog.Name {
			item.Type = MigrationRename
			item.FromName = conf.RenameFrom
		} else if conf.ReIngest == catalog.Name {
			item.Type = MigrationUpdate
		} else {
			// TODO: cache repos
			repoStat, _ := s.Repo.Stat(ctx, s.RepoId, repoPath)
			if repoStat.LastUpdated.After(s.LastMigration) {
				// assume creation until we see a catalog object
				item.Type = MigrationCreate
			} else {
				item.Type = MigrationNoChange
			}
		}

		migrationMap[catalog.Name] = item
	}

	return migrationMap, nil
}

func (s *Service) collectMigrationItems(
	ctx context.Context,
	conf MigrationConfig,
	migrationMap map[string]*MigrationItem,
) []*MigrationItem {
	var migrations []*MigrationItem

	catalogObjs := s.Catalog.FindObjects(ctx, s.InstId, drivers.CatalogObjectTypeUnspecified)
	for _, catalogObj := range catalogObjs {
		apiCatalog, err := catalogObjectToPB(catalogObj)
		if err != nil {
			// TODO
			continue
		}

		item, ok := migrationMap[apiCatalog.Name]
		if !ok {
			if apiCatalog.Name != conf.RenameFrom {
				// no repo present anymore. delete
				item = &MigrationItem{
					CatalogInFile: apiCatalog,
					Type:          MigrationDelete,
				}
			}
			// TODO: handle direct file renames
		} else {
			// update store catalog
			item.CatalogInStore = apiCatalog
			if item.Type == MigrationCreate {
				item.Type = MigrationUpdate
			}
			delete(migrationMap, apiCatalog.Name)
		}

		if item != nil {
			migrations = append(migrations, item)
		}
	}

	for _, migration := range migrationMap {
		migrations = append(migrations, migration)
	}

	return migrations
}

func (s *Service) runMigrationItems(
	ctx context.Context,
	conf MigrationConfig,
	migrations []*MigrationItem,
	result *MigrationResult,
) error {
	for _, migration := range migrations {
		var err error
		switch migration.Type {
		case MigrationCreate:
			err = s.createInStore(ctx, migration)
			result.AddedObjects = append(result.AddedObjects, migration.CatalogInFile)
		case MigrationRename:
			err = s.renameInStore(ctx, migration)
			result.UpdatedObjects = append(result.UpdatedObjects, migration.CatalogInFile)
		case MigrationUpdate:
			err = s.updateInStore(ctx, migration)
			result.UpdatedObjects = append(result.UpdatedObjects, migration.CatalogInFile)
		case MigrationDelete:
			err = s.deleteInStore(ctx, migration)
			result.DroppedObjects = append(result.DroppedObjects, migration.CatalogInFile)
		}
		if err != nil {
			result.ArtifactErrors = append(result.ArtifactErrors, ArtifactError{
				Error: err,
				Path:  migration.Path,
			})
			if !conf.BestEffort {
				return err
			}
		}
	}

	return nil
}

func (s *Service) createInStore(ctx context.Context, item *MigrationItem) error {
	err := artifacts.Write(ctx, s.Repo, s.RepoId, item.CatalogInFile)
	if err != nil {
		return err
	}
	err = migrator.Create(ctx, s.Olap, item.CatalogInFile)
	if err != nil {
		return err
	}
	catalog, err := s.updateCatalogObject(ctx, item)
	if err != nil {
		return err
	}
	return s.Catalog.CreateObject(ctx, s.InstId, catalog)
}

func (s *Service) renameInStore(ctx context.Context, item *MigrationItem) error {
	err := migrator.Rename(ctx, s.Olap, item.FromName, item.CatalogInFile)
	if err != nil {
		return err
	}
	err = s.Catalog.DeleteObject(ctx, s.InstId, item.FromName)
	catalog, err := s.updateCatalogObject(ctx, item)
	if err != nil {
		return err
	}
	return s.Catalog.CreateObject(ctx, s.InstId, catalog)
}

func (s *Service) updateInStore(ctx context.Context, item *MigrationItem) error {
	err := artifacts.Write(ctx, s.Repo, s.RepoId, item.CatalogInFile)
	if err != nil {
		return err
	}
	err = migrator.Update(ctx, s.Olap, item.CatalogInFile)
	if err != nil {
		return err
	}
	catalog, err := s.updateCatalogObject(ctx, item)
	if err != nil {
		return err
	}
	return s.Catalog.UpdateObject(ctx, s.InstId, catalog)
}

func (s *Service) deleteInStore(ctx context.Context, item *MigrationItem) error {
	return migrator.Delete(ctx, s.Olap, item.CatalogInFile)
}

func (s *Service) updateCatalogObject(ctx context.Context, item *MigrationItem) (*drivers.CatalogObject, error) {
	repoStat, err := s.Repo.Stat(ctx, s.RepoId, item.Path)
	if err != nil {
		return nil, err
	}
	catalog, err := pbToCatalogObject(item.CatalogInFile)
	if err != nil {
		return nil, err
	}

	catalog.UpdatedOn = repoStat.LastUpdated
	catalog.RefreshedOn = time.Now()
	return catalog, nil
}
