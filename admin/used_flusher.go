package admin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

type usedFlusher struct {
	deployments map[string]bool
	lock        sync.Mutex
	db          database.DB
	logger      zap.Logger
}

func newUsedFlusher(logger zap.Logger, db database.DB) *usedFlusher {
	return &usedFlusher{
		deployments: make(map[string]bool),
		db:          db,
		logger:      logger,
	}
}

func (u *usedFlusher) UpdateDeplTS(ctx context.Context, deplID string) {
	u.lock.Lock()
	defer u.lock.Unlock()

	u.deployments[deplID] = true
}

func (u *usedFlusher) LastUsedFlusher(ctx context.Context) {
	ticker := time.NewTicker(20 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				u.logger.Info(`Inside LastusedFluser`, zap.Int("no of deployments", len(u.deployments)), observability.ZapCtx(ctx))
				if len(u.deployments) > 0 {
					err := u.updateDeplToDB(ctx)
					if err != nil {
						fmt.Printf("Error while flush update timestamp map into db, error: %v", err)
					}
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (u *usedFlusher) updateDeplToDB(ctx context.Context) error {
	u.lock.Lock()
	defer u.lock.Unlock()

	deplIds := make([]string, 0, len(u.deployments))
	for k := range u.deployments {
		deplIds = append(deplIds, k)
	}

	_, err := u.db.UpdateDeploymentUsedOn(ctx, deplIds)
	u.logger.Info(`Updated deployment status`, zap.Strings("Depl Ids", deplIds), observability.ZapCtx(ctx))

	u.deployments = make(map[string]bool)
	return err
}

func (u *usedFlusher) Close() error {
	return u.db.Close()
}
