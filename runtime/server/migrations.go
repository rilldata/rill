package server

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog"
	"github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Migrate implements RuntimeService
func (s *Server) Migrate(ctx context.Context, req *api.MigrateRequest) (*api.MigrateResponse, error) {
	service, err := s.serviceCache.createCatalogService(ctx, s, req.InstanceId, req.RepoId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := service.Migrate(ctx, catalog.MigrationConfig{
		DryRun:       req.Dry,
		Strict:       req.Strict,
		ChangedPaths: req.ChangedPaths,
	})
	if err != nil {
		return nil, err
	}

	return &api.MigrateResponse{
		Errors:        resp.Errors,
		AffectedPaths: resp.AffectedPaths,
	}, nil
}

// MigrateSingle implements RuntimeService
// NOTE: Everything here is an initial implementation with many flaws.
func (s *Server) MigrateSingle(ctx context.Context, req *api.MigrateSingleRequest) (*api.MigrateSingleResponse, error) {
	// TODO: Handle all kinds of objects, not just sources
	return s.migrateSingleSource(ctx, req)
}

// MigrateDelete implements RuntimeService
func (s *Server) MigrateDelete(ctx context.Context, req *api.MigrateDeleteRequest) (*api.MigrateDeleteResponse, error) {
	// Get instance
	registry, _ := s.metastore.RegistryStore()
	inst, found := registry.FindInstance(ctx, req.InstanceId)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "instance not found")
	}

	// Get catalog
	catalog, err := s.openCatalog(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get object
	obj, found := catalog.FindObject(ctx, req.InstanceId, req.Name)
	if !found {
		return nil, status.Errorf(codes.InvalidArgument, "object not found")
	}

	// Delete from underlying if applicable
	switch obj.Type {
	case drivers.CatalogObjectTypeSource:
		// Get OLAP
		conn, err := s.connCache.openAndMigrate(ctx, inst.ID, inst.Driver, inst.DSN)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
		olap, _ := conn.OLAPStore()

		// Drop table with source name
		rows, err := olap.Execute(ctx, &drivers.Statement{Query: fmt.Sprintf("DROP TABLE %s", obj.Name)})
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
		if err = rows.Close(); err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
	case drivers.CatalogObjectTypeTable:
		// Don't allow deletion of tables created directly in DB
		return nil, status.Error(codes.InvalidArgument, "can not delete unmanaged table")
	case drivers.CatalogObjectTypeMetricsView:
		// Nothing to do
	default:
		panic(fmt.Errorf("unhandled catalog object type: %v", obj.Type))
	}

	// Remove from catalog
	err = catalog.DeleteObject(ctx, req.InstanceId, obj.Name)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "could not delete object: %s", err.Error())
	}

	// Reset catalog cache
	s.catalogCache.reset(req.InstanceId)

	return &api.MigrateDeleteResponse{}, nil
}

// PutFileAndMigrate implements RuntimeService
func (s *Server) PutFileAndMigrate(ctx context.Context, req *api.PutFileAndMigrateRequest) (*api.PutFileAndMigrateResponse, error) {
	_, err := s.PutFile(ctx, &api.PutFileRequest{
		RepoId:     req.RepoId,
		Path:       req.Path,
		Blob:       req.Blob,
		Create:     req.Create,
		CreateOnly: req.CreateOnly,
		Delete:     req.Delete,
	})
	if err != nil {
		return nil, err
	}
	migrateResp, err := s.Migrate(ctx, &api.MigrateRequest{
		InstanceId:   req.InstanceId,
		RepoId:       req.RepoId,
		ChangedPaths: []string{req.Path},
		Dry:          false,
		Strict:       false,
	})
	if err != nil {
		return nil, err
	}
	return &api.PutFileAndMigrateResponse{
		Errors:        migrateResp.Errors,
		AffectedPaths: migrateResp.AffectedPaths,
	}, nil
}

// NOTE: This is an initial migration implementation with several flaws.
func (s *Server) migrateSingleSource(ctx context.Context, req *api.MigrateSingleRequest) (*api.MigrateSingleResponse, error) {
	// Parse SQL
	source, err := sources.SqlToSource(req.Sql)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get instance
	registry, _ := s.metastore.RegistryStore()
	inst, found := registry.FindInstance(ctx, req.InstanceId)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "instance not found")
	}

	// Get catalog
	catalog, err := s.openCatalog(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get olap
	conn, err := s.connCache.openAndMigrate(ctx, inst.ID, inst.Driver, inst.DSN)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	olap, _ := conn.OLAPStore()

	// Get existing object with name and check it's a source
	existingObj, existingFound := catalog.FindObject(ctx, req.InstanceId, source.Name)
	if existingFound && !req.CreateOrReplace {
		return nil, status.Errorf(codes.InvalidArgument, "an existing object with name '%s' already exists (consider setting `create_or_replace=true`)", existingObj.Name)
	}
	if existingFound && existingObj.Type != drivers.CatalogObjectTypeSource {
		return nil, status.Errorf(codes.InvalidArgument, "an object of type '%s' already exists with name '%s'", existingObj.Type, existingObj.Name)
	}

	// Get object to rename and check it's a valid rename op
	var renameObj *drivers.CatalogObject
	var renameFound bool
	var renameAndReingest bool
	if req.RenameFrom != "" {
		// Check that we're not renaming to a name that's already taken
		if existingFound {
			return nil, status.Errorf(codes.InvalidArgument, "cannot rename '%s' to '%s' because a source with that name already exists", req.RenameFrom, source.Name)
		}

		// Get the object to rename
		renameObj, renameFound = catalog.FindObject(ctx, req.InstanceId, req.RenameFrom)
		if !renameFound {
			return nil, status.Errorf(codes.InvalidArgument, "could not find existing object named '%s' to rename", req.RenameFrom)
		}
		if renameObj.Type != drivers.CatalogObjectTypeSource {
			return nil, status.Errorf(codes.InvalidArgument, "cannot rename object '%s' because it is not a source", req.RenameFrom)
		}

		// Check whether the properties for the new object are different (i.e. whether to re-ingest or just rename)
		renameSource, err := sources.SqlToSource(renameObj.SQL)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "could not parse existing sql: %s", err.Error())
		}
		renameAndReingest = !source.PropertiesEquals(renameSource)
	}

	// Stop execution now if it's just a dry run
	if req.DryRun {
		return &api.MigrateSingleResponse{}, nil
	}

	// Create the object to save
	newObj := &drivers.CatalogObject{
		Name:        source.Name,
		Type:        drivers.CatalogObjectTypeSource,
		SQL:         req.Sql,
		RefreshedOn: time.Now(),
	}

	// We now have several cases to handle
	if !existingFound && !renameFound {
		// Just ingest and save object
		err := olap.Ingest(ctx, source)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		st, err := olap.InformationSchema().Lookup(ctx, source.Name)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "couldn't detect schema of ingested source: %s", err.Error())
		}
		newObj.Schema = st.Schema

		err = catalog.CreateObject(ctx, req.InstanceId, newObj)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "error: could not insert source into catalog: %s (warning: watch out for corruptions)", err.Error())
		}
	} else if existingFound && !renameFound {
		// Reingest and then update object
		err := olap.Ingest(ctx, source)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		st, err := olap.InformationSchema().Lookup(ctx, source.Name)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "couldn't detect schema of ingested source: %s", err.Error())
		}
		newObj.Schema = st.Schema

		err = catalog.UpdateObject(ctx, req.InstanceId, newObj)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "error: could not update source in catalog: %s (warning: watch out for corruptions)", err.Error())
		}
	} else if renameFound && !renameAndReingest { // earlier check ensures !existingFound
		// Just create the new object, drop the old one, then rename in OLAP
		err = catalog.CreateObject(ctx, req.InstanceId, newObj)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "error: could not insert source into catalog: %s (warning: watch out for corruptions)", err.Error())
		}

		err = catalog.DeleteObject(ctx, req.InstanceId, renameObj.Name)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}

		rows, err := olap.Execute(ctx, &drivers.Statement{Query: fmt.Sprintf("ALTER TABLE %s RENAME TO %s", renameObj.Name, newObj.Name), Priority: 100})
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
		rows.Close()
	} else if renameFound && renameAndReingest { // earlier check ensures !existingFound
		// Reingest and save object, then drop old
		err := olap.Ingest(ctx, source)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		st, err := olap.InformationSchema().Lookup(ctx, source.Name)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "couldn't detect schema of ingested source: %s", err.Error())
		}
		newObj.Schema = st.Schema

		err = catalog.CreateObject(ctx, req.InstanceId, newObj)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "error: could not insert source into catalog: %s (warning: watch out for corruptions)", err.Error())
		}

		err = catalog.DeleteObject(ctx, req.InstanceId, renameObj.Name)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}

		rows, err := olap.Execute(ctx, &drivers.Statement{Query: fmt.Sprintf("DROP TABLE IF EXISTS %s", renameObj.Name), Priority: 100})
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
		rows.Close()
	}

	// Reset catalog cache
	s.catalogCache.reset(req.InstanceId)

	// Done
	return &api.MigrateSingleResponse{}, nil
}
