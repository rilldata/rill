package river

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

const tokenRetentionPeriod = time.Hour * 24 * 180 // 180 days

// DeleteUnusedUserTokenArgs and Worker for user tokens only

type DeleteUnusedUserTokenArgs struct{}

func (DeleteUnusedUserTokenArgs) Kind() string { return "delete_unused_user_tokens" }

type DeleteUnusedUserTokenWorker struct {
	river.WorkerDefaults[DeleteUnusedUserTokenArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// NewDeleteUnusedUserTokenWorker creates a new worker that deletes unused user tokens.
func (w *DeleteUnusedUserTokenWorker) Work(ctx context.Context, job *river.Job[DeleteUnusedUserTokenArgs]) error {
	err := w.admin.DB.DeleteInactiveUserAuthTokens(ctx, tokenRetentionPeriod)
	if err != nil {
		return fmt.Errorf("failed to delete unused user tokens: %w", err)
	}
	return nil
}

// DeleteUnusedServiceTokenArgs and Worker for service tokens only

type DeleteUnusedServiceTokenArgs struct{}

func (DeleteUnusedServiceTokenArgs) Kind() string { return "delete_unused_service_tokens" }

type DeleteUnusedServiceTokenWorker struct {
	river.WorkerDefaults[DeleteUnusedServiceTokenArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// NewDeleteUnusedServiceTokenWorker creates a new worker that deletes unused service tokens.
func (w *DeleteUnusedServiceTokenWorker) Work(ctx context.Context, job *river.Job[DeleteUnusedServiceTokenArgs]) error {
	err := w.admin.DB.DeleteInactiveServiceAuthTokens(ctx, tokenRetentionPeriod)
	if err != nil {
		return fmt.Errorf("failed to delete unused service tokens: %w", err)
	}
	return nil
}
