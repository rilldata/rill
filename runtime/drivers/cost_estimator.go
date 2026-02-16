package drivers

import "context"

// CostEstimate represents the estimated cost of executing a query.
// Drivers that support dry-run estimation (e.g., BigQuery, Snowflake) populate
// BytesScanned and optionally EstimatedCostUSD. If the driver does not support
// cost estimation, Supported will be false.
type CostEstimate struct {
	// BytesScanned is the estimated number of bytes the query will scan.
	BytesScanned int64
	// EstimatedCostUSD is the estimated monetary cost in USD, if available.
	// It is nil when the driver cannot translate bytes scanned into a dollar amount.
	EstimatedCostUSD *float64
	// Supported indicates whether the driver actually supports cost estimation.
	// When false, the other fields should be ignored.
	Supported bool
}

// CostEstimator is an optional interface that OLAP drivers may implement to
// provide dry-run query cost estimation before actual execution.
//
// Not all drivers support this capability. Callers should use a type assertion
// to check whether a driver handle implements CostEstimator:
//
//	if ce, ok := olapHandle.(drivers.CostEstimator); ok {
//	    estimate, err := ce.EstimateQueryCost(ctx, sql)
//	    // ...
//	}
//
// Drivers that do NOT support cost estimation (e.g., DuckDB) should simply not
// implement this interface. The query execution layer will skip estimation in
// that case.
type CostEstimator interface {
	// EstimateQueryCost performs a dry-run of the given SQL statement and returns
	// an estimate of the resources it would consume. The query is not actually
	// executed.
	//
	// If the driver does not support estimation for a particular query (e.g.,
	// DDL statements), it should return a CostEstimate with Supported set to
	// false and a nil error.
	//
	// Errors are returned only for unexpected failures (network issues, invalid
	// SQL, etc.), not for unsupported estimation.
	EstimateQueryCost(ctx context.Context, sql string) (*CostEstimate, error)
}
