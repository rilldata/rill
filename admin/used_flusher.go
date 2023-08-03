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
	db            database.DB
	logger        *zap.Logger
	mu            sync.Mutex
	deployments   map[string]bool
	users         map[string]bool
	userTokens    map[string]bool
	serviceTokens map[string]bool
	ctx           context.Context
	cancel        context.CancelFunc
	flushWg       sync.WaitGroup
}

func newUsedFlusher(logger *zap.Logger, db database.DB) *usedFlusher {
	ctx, cancel := context.WithCancel(context.Background())

	used := &usedFlusher{
		db:            db,
		logger:        logger,
		deployments:   make(map[string]bool),
		users:         make(map[string]bool),
		userTokens:    make(map[string]bool),
		serviceTokens: make(map[string]bool),
		ctx:           ctx,
		cancel:        cancel,
	}
	go used.runBackground()

	return used
}

func (u *usedFlusher) Deployment(id string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.deployments[id] = true
}

func (u *usedFlusher) User(id string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.users[id] = true
}

func (u *usedFlusher) UserTokens(id string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.userTokens[id] = true
}

func (u *usedFlusher) ServiceTokens(id string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.serviceTokens[id] = true
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

	users := u.users
	u.users = make(map[string]bool)

	userTokens := u.userTokens
	u.userTokens = make(map[string]bool)

	serviceTokens := u.serviceTokens
	u.serviceTokens = make(map[string]bool)
	u.mu.Unlock()

	// Helper function to perform the flushing of used_on to the database.
	flushToDB := func(data map[string]bool, updateFn func(ctx context.Context, ids []string) error, logMsg string) {
		if len(data) > 0 {
			u.logger.Info("flushing used_on to db", zap.Int(logMsg, len(data)))

			ctx, cancel := context.WithTimeout(context.Background(), flushTimeout)
			defer cancel()

			ids := make([]string, 0, len(data))
			for k := range data {
				ids = append(ids, k)
			}

			err := updateFn(ctx, ids)
			if err != nil {
				u.logger.Error("flushing used_on failed", zap.Error(err), zap.Strings(logMsg, ids), observability.ZapCtx(ctx))
			}

			u.logger.Info("flushed used_on to db", zap.Int(logMsg, len(data)))
		}
	}

	// Flush deployments
	flushToDB(deployments, u.db.UpdateDeploymentUsedOn, "deployments")

	// Flush users
	flushToDB(users, u.db.UpdateUserActiveOn, "users")

	// Flush user tokens
	flushToDB(userTokens, u.db.UpdateUserAuthTokenUsedOn, "user tokens")

	// Flush service tokens
	flushToDB(serviceTokens, u.db.UpdateServiceAuthTokenUsedOn, "service tokens")
}
