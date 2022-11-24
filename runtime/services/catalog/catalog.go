package catalog

import (
	"context"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/dag"
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
	dag           *dag.DAG
	// used to get path when we only have name. happens when we get name from DAG
	// TODO: should we add path to the DAG instead
	NameToPath map[string]string
	// used to get last logged name when parsing fails
	PathToName map[string]string
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

		dag:        dag.NewDAG(),
		NameToPath: make(map[string]string),
		PathToName: make(map[string]string),
	}
}

func (s *Service) ListObjects(ctx context.Context, typ runtimev1.ObjectType) ([]*runtimev1.CatalogEntry, error) {
	objs := s.Catalog.FindEntries(ctx, s.InstId, pbToObjectType(typ))
	pbs := make([]*runtimev1.CatalogEntry, len(objs))
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
) (*runtimev1.CatalogEntry, error) {
	obj, found := s.Catalog.FindEntry(ctx, s.InstId, name)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "object not found")
	}

	pb, err := catalogObjectToPB(obj)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return pb, nil
}
