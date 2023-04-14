package catalog

import (
	"context"
	"sync"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/dag"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type Service struct {
	Catalog       drivers.CatalogStore
	Repo          drivers.RepoStore
	Olap          drivers.OLAPStore
	RegistryStore drivers.RegistryStore
	InstID        string

	// temporary information. should this be persisted into olap?
	// LastMigration stores the last time migrate was run. Used to filter out repos that didnt change since this time
	LastMigration time.Time
	dag           *dag.DAG
	// used to get path when we only have name. happens when we get name from DAG
	// TODO: should we add path to the DAG instead
	NameToPath map[string]string

	logger      *otelzap.Logger
	hasMigrated bool
	lock        sync.Mutex
}

func NewService(
	catalog drivers.CatalogStore,
	repo drivers.RepoStore,
	olap drivers.OLAPStore,
	registry drivers.RegistryStore,
	instID string,
	logger *otelzap.Logger,
) *Service {
	if logger == nil {
		logger = otelzap.New(zap.NewNop())
	}
	return &Service{
		Catalog:       catalog,
		Repo:          repo,
		Olap:          olap,
		RegistryStore: registry,
		InstID:        instID,

		dag:        dag.NewDAG(),
		NameToPath: make(map[string]string),

		logger: logger,
	}
}

func (s *Service) FindEntries(ctx context.Context, typ drivers.ObjectType) []*drivers.CatalogEntry {
	entries := s.Catalog.FindEntries(ctx, s.InstID, typ)
	for _, entry := range entries {
		s.fillDAGInEntry(entry)
	}
	return entries
}

func (s *Service) FindEntry(ctx context.Context, name string) (*drivers.CatalogEntry, bool) {
	entry, ok := s.Catalog.FindEntry(ctx, s.InstID, name)
	if ok {
		s.fillDAGInEntry(entry)
	}
	return entry, ok
}

func (s *Service) fillDAGInEntry(entry *drivers.CatalogEntry) {
	entry.Children = s.dag.GetChildren(normalizeName(entry.Name))
	entry.Parents = s.dag.GetParents(normalizeName(entry.Name))
}
