package worker

import (
	"context"

	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

func (w *Worker) deleteExpiredTokens(ctx context.Context) error {
	expiredTokens, err := w.admin.DB.FindExpiredUserAuthTokens(ctx)
	if err != nil {
		return err
	}

	for _, token := range expiredTokens {
		w.logger.Info(`deleting expired token`, zap.String("id", token.ID), zap.Time("expires_on", *token.ExpiresOn), observability.ZapCtx(ctx))
		err = w.admin.DB.DeleteUserAuthToken(ctx, token.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
