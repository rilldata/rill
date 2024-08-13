package riverutils

import (
	"database/sql"

	"github.com/rilldata/rill/admin/database"
	"github.com/riverqueue/river"
)

// InsertOnlyRiverClient is a river client that only supports inserting jobs, actual river worker is started in the worker service
var InsertOnlyRiverClient *river.Client[*sql.Tx]

// AddBillingErrorArgs is the arguments for AddBillingErrorWorker
type AddBillingErrorArgs struct {
	CustomerID string
	ErrorType  database.BillingErrorType
	Metadata   map[string]string
}

func (AddBillingErrorArgs) Kind() string { return "add_billing_error" }

type ChargeSuccessArgs struct {
	CustomerID string `json:"customer_id"`
	Metadata   map[string]string
}

func (ChargeSuccessArgs) Kind() string { return "charge_success" }
