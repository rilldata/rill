package catalog

import (
	"context"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/dag"
	"go.uber.org/zap"
)

type Service struct {
	Catalog drivers.CatalogStore
	Repo    drivers.RepoStore
	Olap    drivers.OLAPStore
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

	logger *zap.Logger
}

func NewService(
	catalog drivers.CatalogStore,
	repo drivers.RepoStore,
	olap drivers.OLAPStore,
	instId string,
	logger *zap.Logger,
) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		Catalog: catalog,
		Repo:    repo,
		Olap:    olap,
		InstId:  instId,

		dag:        dag.NewDAG(),
		NameToPath: make(map[string]string),
		PathToName: make(map[string]string),

		logger: logger,
	}
}

func (s *Service) FindEntries(ctx context.Context, typ drivers.ObjectType) []*drivers.CatalogEntry {
	return s.Catalog.FindEntries(ctx, s.InstId, typ)
}

func (s *Service) FindEntry(ctx context.Context, name string) (*drivers.CatalogEntry, bool) {
	return s.Catalog.FindEntry(ctx, s.InstId, name)
}
