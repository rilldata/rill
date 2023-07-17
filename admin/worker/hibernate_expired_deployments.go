package worker

import "context"

func (w *Worker) hibernateExpiredDeployments(ctx context.Context) error {
	return w.admin.HibernateDeployments(ctx)
}
