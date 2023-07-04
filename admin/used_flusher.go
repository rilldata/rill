package admin

import (
	"context"
	"sync"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

const (
	flushInterval = 30 * time.Second
	flushTimeout  = 15 * time.Second
)

type usedFlusher struct {
	db          database.DB
	logger      *zap.Logger
	mu          sync.Mutex
	deployments map[string]bool
	ctx         context.Context
	cancel      context.CancelFunc
	flushWg     sync.WaitGroup
}

func newUsedFlusher(logger *zap.Logger, db database.DB) *usedFlusher {
	ctx, cancel := context.WithCancel(context.Background())

	used := &usedFlusher{
		db:          db,
		logger:      logger,
		deployments: make(map[string]bool),
		ctx:         ctx,
		cancel:      cancel,
	}
	go used.runBackground()

	return used
}

func (u *usedFlusher) Deployment(id string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.deployments[id] = true
}

func (u *usedFlusher) Close() {
	u.cancel()
	u.flush()
	u.flushWg.Wait()
}

func (u *usedFlusher) runBackground() {
	ticker := time.NewTicker(flushInterval)
	for {
		select {
		case <-ticker.C:
			u.flush()
		case <-u.ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (u *usedFlusher) flush() {
	u.flushWg.Add(1)
	defer u.flushWg.Done()

	u.mu.Lock()
	deployments := u.deployments
	u.deployments = make(map[string]bool)
	u.mu.Unlock()

	if len(deployments) > 0 {
		u.logger.Info("flushing used_on to db", zap.Int("deployments", len(deployments)))

		ctx, cancel := context.WithTimeout(context.Background(), flushTimeout)
		defer cancel()

		ids := make([]string, 0, len(deployments))
		for k := range deployments {
			ids = append(ids, k)
		}

		err := u.db.UpdateDeploymentUsedOn(ctx, ids)
		if err != nil {
			u.logger.Error("flushing used_on failed", zap.Error(err), zap.Strings("deployment_ids", ids), observability.ZapCtx(ctx))
		}

		u.logger.Info("flushed used_on to db", zap.Int("deployments", len(deployments)))
	}
}
