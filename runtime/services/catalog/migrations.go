package catalog

import (
	"context"
	"time"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
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

	// TODO: if the repo folder is source controlled we should leverage it to find changes
	// TODO: ListRecursive needs some kind of cache or optimisation
	repoPaths, err := s.Repo.ListRecursive(ctx, s.RepoId)
	if err != nil {
		return result, err
	}
	migrationMap := make(map[string]*MigrationItem)
	var migrations []*MigrationItem
	now := time.Now()
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
	// There could be changes to repo between last migration and while the repos are being read
	// hence we log LastMigration as the time before we start reading the repos and assign here
	s.LastMigration = now

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

		migrations = append(migrations, item)
	}

	for _, migration := range migrationMap {
		migrations = append(migrations, migration)
	}

	if conf.DryRun {
		return result, nil
	}

	for _, migration := range migrations {
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
				return result, err
			}
		}
	}

	return result, nil
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
