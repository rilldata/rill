package river

import (
	"context"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

// connectorOAuthClientRetention is how long a connector OAuth client may go unused before it is deleted together with its tokens.
// Clients stop being used when the connector's OAuth configuration changes, the connector is removed, or its branch deployment goes away.
const connectorOAuthClientRetention = 30 * 24 * time.Hour

type DeleteUnusedConnectorOAuthClientsArgs struct{}

func (DeleteUnusedConnectorOAuthClientsArgs) Kind() string {
	return "delete_unused_connector_oauth_clients"
}

type DeleteUnusedConnectorOAuthClientsWorker struct {
	river.WorkerDefaults[DeleteUnusedConnectorOAuthClientsArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *DeleteUnusedConnectorOAuthClientsWorker) Work(ctx context.Context, job *river.Job[DeleteUnusedConnectorOAuthClientsArgs]) error {
	return w.admin.DB.DeleteUnusedConnectorOAuthClients(ctx, time.Now().Add(-connectorOAuthClientRetention))
}
