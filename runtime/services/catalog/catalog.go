package catalog

import (
	"context"
	"time"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	dag2 "github.com/rilldata/rill/runtime/pkg/dag"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	Catalog drivers.CatalogStore
	Repo    drivers.RepoStore
	Olap    drivers.OLAPStore
	RepoId  string
	InstId  string

	// temporary information. should this be persisted into olap?
	// LastMigration stores the last time migrate was run. Used to filter out repos that didnt change since this time
	LastMigration time.Time
	dag           *dag2.DAG
	NameToPath    map[string]string
	PathToName    map[string]string
}

func NewService(
	catalog drivers.CatalogStore,
	repo drivers.RepoStore,
	olap drivers.OLAPStore,
	repoId string,
	instId string,
) *Service {
	return &Service{
		Catalog: catalog,
		Repo:    repo,
		Olap:    olap,
		RepoId:  repoId,
		InstId:  instId,

		dag:        dag2.NewDAG(),
		NameToPath: make(map[string]string),
		PathToName: make(map[string]string),
	}
}

func (s *Service) ListObjects(
	ctx context.Context,
) ([]*api.CatalogObject, error) {
	objs := s.Catalog.FindObjects(ctx, s.InstId, drivers.CatalogObjectTypeUnspecified)
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
