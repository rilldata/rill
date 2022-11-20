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

func (c *servicesCache) createCatalogService(ctx context.Context, s *Server, instId string, repoId string) (*catalog.Service, error) {
	// TODO 1: opening a driver shouldn't take too long but we should still have an instance specific lock
	// TODO 2: This is a cache on a cache, which may lead to undefined behavior
	// TODO 3: This relies on a one-to-one coupling between instances and repos, which we could formalize by making repos part of instances

	c.lock.Lock()
	defer c.lock.Unlock()

	// right now there is 1-1 mapping from instance to repo.
	// TODO: support both instance and repo in this key
	key := instId + repoId

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

	repo, ok := registry.FindRepo(ctx, repoId)
	if !ok && repoId != "" {
		return nil, status.Errorf(codes.InvalidArgument, "repo '%s' not found", repoId)
	}

	var repoStore drivers.RepoStore
	if ok {
		repoConn, err := drivers.Open(repo.Driver, repo.DSN)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		repoStore, ok = repoConn.RepoStore()
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "repo '%s' is not a valid repo store", repoId)
		}
	}

	service = catalog.NewService(catalogStore, repoStore, olap, repoId, instId)
	c.catalogServices[key] = service
	return service, nil
}
