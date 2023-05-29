package worker

import (
	"context"

	"go.uber.org/zap"
)

func (w *Worker) deleteExpiredTokens(ctx context.Context) error {
	expiredTokens, err := w.admin.DB.FindExpiredAuthTokens(ctx)
	if err != nil {
		return err
	}

	for _, token := range expiredTokens {
		w.logger.Info(`deleting expired tokens`, zap.String("id", token.ID), zap.String("expired_ts", token.ExpirationTS.String()))
		err = w.admin.DB.DeleteUserAuthToken(ctx, token.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
