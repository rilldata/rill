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
	db               database.DB
	logger           *zap.Logger
	mu               sync.Mutex
	deployments      map[string]bool
	users            map[string]bool
	service          map[string]bool
	userTokens       map[string]bool
	serviceTokens    map[string]bool
	deploymentTokens map[string]bool
	magicAuthTokens  map[string]bool
	client           map[string]bool
	ctx              context.Context
	cancel           context.CancelFunc
	flushWg          sync.WaitGroup
}

func newUsedFlusher(logger *zap.Logger, db database.DB) *usedFlusher {
	ctx, cancel := context.WithCancel(context.Background())

	used := &usedFlusher{
		db:               db,
		logger:           logger,
		deployments:      make(map[string]bool),
		users:            make(map[string]bool),
		service:          make(map[string]bool),
		userTokens:       make(map[string]bool),
		serviceTokens:    make(map[string]bool),
		deploymentTokens: make(map[string]bool),
		magicAuthTokens:  make(map[string]bool),
		client:           make(map[string]bool),
		ctx:              ctx,
		cancel:           cancel,
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

func (u *usedFlusher) Service(id string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.service[id] = true
}

func (u *usedFlusher) UserToken(id string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.userTokens[id] = true
}

func (u *usedFlusher) ServiceToken(id string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.serviceTokens[id] = true
}

func (u *usedFlusher) DeploymentToken(id string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.deploymentTokens[id] = true
}

func (u *usedFlusher) MagicAuthToken(id string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.magicAuthTokens[id] = true
}

func (u *usedFlusher) Client(id string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.client[id] = true
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
	var deployments map[string]bool
	if len(u.deployments) > 0 {
		deployments = u.deployments
		u.deployments = make(map[string]bool)
	}

	var users map[string]bool
	if len(u.users) > 0 {
		users = u.users
		u.users = make(map[string]bool)
	}

	var service map[string]bool
	if len(u.service) > 0 {
		service = u.service
		u.service = make(map[string]bool)
	}

	var userTokens map[string]bool
	if len(u.userTokens) > 0 {
		userTokens = u.userTokens
		u.userTokens = make(map[string]bool)
	}

	var serviceTokens map[string]bool
	if len(u.serviceTokens) > 0 {
		serviceTokens = u.serviceTokens
		u.serviceTokens = make(map[string]bool)
	}

	var deploymentTokens map[string]bool
	if len(u.deploymentTokens) > 0 {
		deploymentTokens = u.deploymentTokens
		u.deploymentTokens = make(map[string]bool)
	}

	var magicAuthTokens map[string]bool
	if len(u.magicAuthTokens) > 0 {
		magicAuthTokens = u.magicAuthTokens
		u.magicAuthTokens = make(map[string]bool)
	}

	var client map[string]bool
	if len(u.client) > 0 {
		client = u.client
		u.client = make(map[string]bool)
	}
	u.mu.Unlock()

	// Flush deployments
	u.flushToDB(deployments, u.db.UpdateDeploymentUsedOn, "deployments")

	// Flush users
	u.flushToDB(users, u.db.UpdateUserActiveOn, "users")

	// Flush services
	u.flushToDB(service, u.db.UpdateServiceActiveOn, "services")

	// Flush user tokens
	u.flushToDB(userTokens, u.db.UpdateUserAuthTokenUsedOn, "user tokens")

	// Flush service tokens
	u.flushToDB(serviceTokens, u.db.UpdateServiceAuthTokenUsedOn, "service tokens")

	// Flush deployment tokens
	u.flushToDB(deploymentTokens, u.db.UpdateDeploymentAuthTokenUsedOn, "deployment tokens")

	// Flush magic auth tokens
	u.flushToDB(magicAuthTokens, u.db.UpdateMagicAuthTokenUsedOn, "magic auth tokens")

	// Flush auth clients
	u.flushToDB(client, u.db.UpdateAuthClientUsedOn, "auth clients")
}

// Helper function to perform the flushing of used_on to the database.
func (u *usedFlusher) flushToDB(data map[string]bool, updateFn func(ctx context.Context, ids []string) error, logMsg string) {
	if len(data) > 0 {
		u.logger.Debug("flushing used_on to db", zap.Int(logMsg, len(data)))

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

		u.logger.Debug("flushed used_on to db", zap.Int(logMsg, len(data)))
	}
}
