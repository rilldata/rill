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

func (c *servicesCache) createCatalogService(ctx context.Context, s *Server, instId string) (*catalog.Service, error) {
	// TODO 1: opening a driver shouldn't take too long but we should still have an instance specific lock
	// TODO 2: This is a cache on a cache, which may lead to undefined behavior

	c.lock.Lock()
	defer c.lock.Unlock()

	key := instId

	service, ok := c.catalogServices[key]
	if ok {
		return service, nil
	}

	registry, _ := s.metastore.RegistryStore()
	inst, err := registry.FindInstance(ctx, instId)
	if err != nil {
		if err == drivers.ErrNotFound {
			return nil, status.Error(codes.InvalidArgument, "instance not found")
		}
		return nil, err
	}

	olapConn, err := s.connCache.openAndMigrate(ctx, instId, inst.OLAPDriver, inst.OLAPDSN)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	olap, _ := olapConn.OLAPStore()

	catalogStore, err := s.openCatalog(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	repoConn, err := drivers.Open(inst.RepoDriver, inst.RepoDSN)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	repoStore, ok := repoConn.RepoStore()
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "instance '%s' doesn't have a valid repo", instId)
	}

	service = catalog.NewService(catalogStore, repoStore, olap, instId)
	c.catalogServices[key] = service
	return service, nil
}
