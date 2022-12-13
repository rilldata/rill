package catalog

import (
	"context"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/dag"
)

type Service struct {
	Catalog drivers.CatalogStore
	Repo    drivers.RepoStore
	Olap    drivers.OLAPStore
	InstID  string

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

func NewService(catalog drivers.CatalogStore, repo drivers.RepoStore, olap drivers.OLAPStore, instID string) *Service {
	return &Service{
		Catalog: catalog,
		Repo:    repo,
		Olap:    olap,
		InstID:  instID,

		dag:        dag.NewDAG(),
		NameToPath: make(map[string]string),
		PathToName: make(map[string]string),
	}
}

func (s *Service) FindEntries(ctx context.Context, typ drivers.ObjectType) []*drivers.CatalogEntry {
	return s.Catalog.FindEntries(ctx, s.InstID, typ)
}

func (s *Service) FindEntry(ctx context.Context, name string) (*drivers.CatalogEntry, bool) {
	return s.Catalog.FindEntry(ctx, s.InstID, name)
}
