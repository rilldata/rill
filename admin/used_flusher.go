package admin

import (
	"context"
	"sync"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

const flushDuration = 20 * time.Second

type usedFlusher struct {
	deployments map[string]bool
	lock        sync.Mutex
	db          database.DB
	logger      *zap.Logger
	ctx         context.Context
	cancel      context.CancelFunc
}

func (u *usedFlusher) Deployment(id string) {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.deployments[id] = true
}

func (u *usedFlusher) Close() {
	u.flush(context.Background())
	u.cancel()
}

func newUsedFlusher(logger *zap.Logger, db database.DB) *usedFlusher {
	ctx, cancel := context.WithCancel(context.Background())

	used := &usedFlusher{
		deployments: make(map[string]bool),
		db:          db,
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
	}
	go used.runBackground()

	return used
}

func (u *usedFlusher) runBackground() {
	ticker := time.NewTicker(flushDuration)
	for {
		select {
		case <-ticker.C:
			u.flush(u.ctx)
		case <-u.ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (u *usedFlusher) flush(ctx context.Context) {
	u.logger.Info(`flush deployments`, zap.Int("no of deployments", len(u.deployments)), observability.ZapCtx(u.ctx))

	if len(u.deployments) > 0 {
		u.lock.Lock()
		deployments := u.deployments
		u.deployments = make(map[string]bool)
		u.lock.Unlock()

		deplIds := make([]string, 0, len(deployments))
		for k := range deployments {
			deplIds = append(deplIds, k)
		}

		_, err := u.db.UpdateDeploymentUsedOn(ctx, deplIds)
		if err != nil {
			u.logger.Error("Error while flush update timestamp map into db", zap.Strings("deployment_ids", deplIds), zap.Error(err), observability.ZapCtx(ctx))
		}

		u.logger.Info(`Updated deployment status`, zap.Strings("Depl Ids", deplIds), observability.ZapCtx(ctx))
	}
}
