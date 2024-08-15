package riverutils

import (
	"database/sql"
	"time"

	"github.com/riverqueue/river"
)

// InsertOnlyRiverClient is a river client that only supports inserting jobs, actual river worker is started in the worker service
var InsertOnlyRiverClient *river.Client[*sql.Tx]

type ChargeFailedArgs struct {
	ID         string // for deduplication
	CustomerID string
	Currency   string
	Amount     int64
	EventTime  time.Time
}

func (ChargeFailedArgs) Kind() string { return "charge_failed" }

type ChargeSuccessArgs struct {
	ID         string // for deduplication
	CustomerID string
	Amount     int64
	Currency   string
	EventTime  time.Time
}

func (ChargeSuccessArgs) Kind() string { return "charge_success" }

type PaymentMethodAdded struct {
	ID          string // for deduplication
	CustomerID  string
	PaymentType string
	EventTime   time.Time
}

func (PaymentMethodAdded) Kind() string { return "payment_method_added" }

type PaymentMethodRemoved struct {
	ID         string // for deduplication
	CustomerID string
	EventTime  time.Time
}

func (PaymentMethodRemoved) Kind() string { return "payment_method_removed" }
