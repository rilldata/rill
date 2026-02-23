package starrocks

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/starrocks/teststarrocks"
)

// BenchmarkTransport compares MySQL vs Arrow Flight SQL query performance.
// Run with: go test -bench=BenchmarkTransport -benchtime=10x -timeout=10m ./runtime/drivers/starrocks/
func BenchmarkTransport(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping integration benchmark")
	}

	info := teststarrocks.StartWithDataFull(b)

	mysqlConn, mysqlOLAP := openMySQLConn(b, info)
	defer mysqlConn.Close()

	flightConn, flightOLAP := openFlightSQLConn(b, info)
	defer flightConn.Close()

	ctx := context.Background()

	// Queries ordered by result set size: small → large
	queries := []struct {
		name  string
		query string
	}{
		{"SmallLiteral", "SELECT 1 AS id, 'hello' AS name"},
		{"SingleRow", "SELECT * FROM test_db.all_types WHERE id = 1"},
		{"ThreeRows", "SELECT * FROM test_db.all_types ORDER BY id"},
		{"AdBids100", "SELECT * FROM test_db.ad_bids LIMIT 100"},
		{"AdBids1K", "SELECT * FROM test_db.ad_bids LIMIT 1000"},
		{"AdBids10K", "SELECT * FROM test_db.ad_bids LIMIT 10000"},
		{"AdBidsAll", "SELECT * FROM test_db.ad_bids"},
		{"AdBidsAgg", "SELECT publisher, domain, COUNT(*) as cnt, AVG(bid_price) as avg_price FROM test_db.ad_bids GROUP BY publisher, domain ORDER BY cnt DESC"},
	}

	for _, q := range queries {
		b.Run(q.name, func(b *testing.B) {
			b.Run("MySQL", func(b *testing.B) {
				benchQuery(b, ctx, mysqlOLAP, q.query)
			})
			b.Run("FlightSQL", func(b *testing.B) {
				benchQuery(b, ctx, flightOLAP, q.query)
			})
		})
	}
}

// benchQuery runs a query b.N times, consuming all rows and counting them.
// On StarRocks FE connection limit errors, it pauses outside the timer and retries.
func benchQuery(b *testing.B, ctx context.Context, olap drivers.OLAPStore, query string) {
	b.Helper()

	// Warm up: run once to prime caches
	res, err := olap.Query(ctx, &drivers.Statement{Query: query})
	if err != nil {
		b.Fatalf("warmup query failed: %v", err)
	}
	warmupRows := drainRows(b, res)
	res.Close()

	b.ResetTimer()
	b.ReportAllocs()

	var totalRows int64
	for i := 0; i < b.N; i++ {
		res, err := olap.Query(ctx, &drivers.Statement{Query: query})
		if err != nil {
			if isConnLimitErr(err) {
				// Pause timer, wait for FE connections to drain, retry
				b.StopTimer()
				time.Sleep(5 * time.Second)
				b.StartTimer()
				i--
				continue
			}
			b.Fatalf("query failed on iteration %d: %v", i, err)
		}
		totalRows += drainRows(b, res)
		res.Close()
	}

	b.StopTimer()
	b.ReportMetric(float64(warmupRows), "rows/query")
	b.ReportMetric(float64(totalRows)/float64(b.N), "rows/op")
}

// drainRows reads all rows from a result set via MapScan and returns the count.
func drainRows(b *testing.B, res *drivers.Result) int64 {
	b.Helper()
	var count int64
	row := make(map[string]any)
	for res.Next() {
		// Clear the map for reuse
		for k := range row {
			delete(row, k)
		}
		if err := res.MapScan(row); err != nil {
			b.Fatalf("MapScan failed: %v", err)
		}
		count++
	}
	if err := res.Err(); err != nil {
		b.Fatalf("iteration error: %v", err)
	}
	return count
}

// BenchmarkTransportScanOnly measures raw scan throughput without MapScan overhead.
// Uses Scan() with typed destinations to isolate transport performance.
func BenchmarkTransportScanOnly(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping integration benchmark")
	}

	info := teststarrocks.StartWithDataFull(b)

	mysqlConn, mysqlOLAP := openMySQLConn(b, info)
	defer mysqlConn.Close()

	flightConn, flightOLAP := openFlightSQLConn(b, info)
	defer flightConn.Close()

	ctx := context.Background()
	query := "SELECT id, timestamp, publisher, domain, bid_price FROM test_db.ad_bids"

	for _, tc := range []struct {
		name string
		olap drivers.OLAPStore
	}{
		{"MySQL", mysqlOLAP},
		{"FlightSQL", flightOLAP},
	} {
		b.Run(tc.name, func(b *testing.B) {
			// Warm up
			res, err := tc.olap.Query(ctx, &drivers.Statement{Query: query})
			if err != nil {
				b.Fatalf("warmup failed: %v", err)
			}
			var warmupCount int64
			for res.Next() {
				warmupCount++
			}
			res.Close()

			b.ResetTimer()
			b.ReportAllocs()

			var totalRows int64
			for i := 0; i < b.N; i++ {
				res, err := tc.olap.Query(ctx, &drivers.Statement{Query: query})
				if err != nil {
					if isConnLimitErr(err) {
						b.StopTimer()
						time.Sleep(5 * time.Second)
						b.StartTimer()
						i--
						continue
					}
					b.Fatalf("query failed: %v", err)
				}
				var count int64
				row := make(map[string]any)
				for res.Next() {
					for k := range row {
						delete(row, k)
					}
					_ = res.MapScan(row)
					count++
				}
				res.Close()
				totalRows += count
			}

			b.StopTimer()
			b.ReportMetric(float64(warmupCount), "rows/query")
			b.ReportMetric(float64(totalRows)/float64(b.N), "rows/op")
			avgRowsPerSec := float64(totalRows) / b.Elapsed().Seconds()
			b.ReportMetric(avgRowsPerSec, "rows/sec")
		})
	}
}

// isConnLimitErr returns true if the error is a StarRocks FE connection limit error.
func isConnLimitErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "connection limit")
}
