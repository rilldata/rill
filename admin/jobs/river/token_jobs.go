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

// DeleteExpiredTokenArgs and Worker for both user and service tokens

type DeleteExpiredTokenArgs struct{}

func (DeleteExpiredTokenArgs) Kind() string { return "delete_expired_token" }

type DeleteExpiredTokenWorker struct {
	river.WorkerDefaults[DeleteExpiredTokenArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *DeleteExpiredTokenWorker) Work(ctx context.Context, job *river.Job[DeleteExpiredTokenArgs]) error {
	// Delete auth tokens that have been expired for more than 24 hours
	retention := 24 * time.Hour
	err := w.admin.DB.DeleteExpiredUserAuthTokens(ctx, retention)
	if err != nil {
		return err
	}
	err = w.admin.DB.DeleteExpiredServiceAuthTokens(ctx, retention)
	if err != nil {
		return err
	}
	err = w.admin.DB.DeleteExpiredDeploymentAuthTokens(ctx, retention)
	if err != nil {
		return err
	}
	err = w.admin.DB.DeleteExpiredMagicAuthTokens(ctx, retention)
	if err != nil {
		return err
	}
	return nil
}
