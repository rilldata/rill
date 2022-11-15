package server

import (
	"context"
	"sync"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type servicesCache struct {
	lock            sync.Mutex
	catalogServices map[string]*catalog.Service
}

func newServicesCache() *servicesCache {
	return &servicesCache{
		catalogServices: make(map[string]*catalog.Service),
	}
}

func (c *servicesCache) createCatalogService(
	ctx context.Context,
	s *Server,
	instId string,
	repoId string,
) (*catalog.Service, error) {
	// TODO: opening a driver shouldn't take too long but we should still have a instance specific lock
	c.lock.Lock()
	defer c.lock.Unlock()

	// right now there is 1-1 mapping from instance to repo.
	// TODO: support both instance and repo in this key
	key := instId

	service, ok := c.catalogServices[key]
	if ok {
		return service, nil
	}

	registry, _ := s.metastore.RegistryStore()
	inst, found := registry.FindInstance(ctx, instId)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "instance not found")
	}

	olapConn, err := s.connCache.openAndMigrate(ctx, instId, inst.Driver, inst.DSN)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	olap, _ := olapConn.OLAPStore()

	catalogStore, err := s.openCatalog(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	repo, _ := registry.FindRepo(ctx, repoId)
	repoConn, err := drivers.Open(repo.Driver, repo.DSN)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	repoStore, _ := repoConn.RepoStore()

	service = catalog.NewService(catalogStore, repoStore, olap, repoId, instId)
	c.catalogServices[key] = service
	return service, nil
}
