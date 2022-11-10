package catalog

import (
	"context"
	"time"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	Catalog drivers.CatalogStore
	RepoId  string
	Repo    drivers.RepoStore
	InstId  string
	Olap    drivers.OLAPStore

	// temporary information. should this be persisted into olap?
	// LastMigration stores the last time migrate was run. Used to filter out repos that didnt change since this time
	LastMigration time.Time
}

func (s *Service) ListObjects(
	ctx context.Context,
) ([]*api.CatalogObject, error) {
	objs := s.Catalog.FindObjects(ctx, s.InstId)
	pbs := make([]*api.CatalogObject, len(objs))
	var err error
	for i, obj := range objs {
		pbs[i], err = catalogObjectToPB(obj)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	return pbs, nil
}

func (s *Service) GetCatalogObject(
	ctx context.Context,
	name string,
) (*api.CatalogObject, error) {
	obj, found := s.Catalog.FindObject(ctx, s.InstId, name)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "object not found")
	}

	pb, err := catalogObjectToPB(obj)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return pb, nil
}

func (s *Service) TriggerRefresh(
	ctx context.Context,
	name string,
) error {
	// Find object
	obj, found := s.Catalog.FindObject(ctx, s.InstId, name)
	if !found {
		return status.Error(codes.InvalidArgument, "object not found")
	}

	switch obj.Type {
	case drivers.CatalogObjectTypeSource:
		// Parse SQL
		source, err := sources.SqlToSource(obj.SQL)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		// Ingest the source
		err = s.Olap.Ingest(ctx, source)
		if err != nil {
			return status.Error(codes.Unknown, err.Error())
		}

		// Update object
		obj.RefreshedOn = time.Now()
		err = s.Catalog.UpdateObject(ctx, s.InstId, obj)

	case drivers.CatalogObjectTypeModel:
		//TODO
	}

	return nil
}
