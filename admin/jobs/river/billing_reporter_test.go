package river

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/metrics"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBillingReporterPaginationBoundariesAndTupleCursor(t *testing.T) {
	// The metrics API uses a six-column cursor. Exact page boundaries must neither
	// skip the first row of the next page nor loop when timestamps are identical.
	tests := []struct {
		rows      int
		wantCalls int
	}{
		{rows: 0, wantCalls: 1},
		{rows: 9_999, wantCalls: 1},
		{rows: 10_000, wantCalls: 2},
		{rows: 10_001, wantCalls: 2},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d rows", tt.rows), func(t *testing.T) {
			worker, db, biller, client := newBillingReporterFixture(makeBillingMetricRows(tt.rows))

			err := worker.Work(t.Context(), nil)
			require.NoError(t, err)
			require.Len(t, client.calls, tt.wantCalls)
			require.Equal(t, tt.rows, countReportedMetric(biller.delivered, "query"))

			if tt.rows >= 10_000 {
				last := client.rows[9_999]
				second := client.calls[1]
				require.Equal(t, last.StartTime, second.afterTime)
				require.Equal(t, last.OrgID, second.afterOrgID)
				require.Equal(t, last.ProjectID, second.afterProjectID)
				require.Equal(t, last.InstanceID, second.afterInstanceID)
				require.Equal(t, last.BillingService, second.afterBillingService)
				require.Equal(t, last.EventName, second.afterEventName)
			}
			if tt.rows == 0 {
				require.True(t, db.checkpoint.IsZero())
			} else {
				require.Equal(t, client.rows[len(client.rows)-1].EndTime, db.checkpoint)
				require.Equal(t, 1, countReportedMetric(biller.delivered, "seats"))
			}
		})
	}
}

func TestBillingReporterUsesSumForCountersAndMaxForGauges(t *testing.T) {
	// Counter metrics conserve all events in a grain, while gauges represent the
	// peak. Swapping these silently over- or under-bills customers.
	rows := makeBillingMetricRows(2)
	rows[0].EventName = "query"
	rows[0].SumValue = 17
	rows[0].MaxValue = 99
	rows[1].EventName = "storage_bytes"
	rows[1].SumValue = 1_000
	rows[1].MaxValue = 250
	worker, _, biller, _ := newBillingReporterFixture(rows)

	require.NoError(t, worker.Work(t.Context(), nil))
	runtimeUsage := findBillingUsageBatch(t, biller.delivered, "query")
	require.Len(t, runtimeUsage, 2)
	require.EqualValues(t, 17, runtimeUsage[0].Value)
	require.EqualValues(t, 250, runtimeUsage[1].Value)
	require.Equal(t, "org-1", runtimeUsage[0].Metadata["org_id"])
}

func TestBillingReporterFailurePathsDoNotAdvanceFinalCheckpoint(t *testing.T) {
	// A checkpoint is a durability claim. Fetch, provider, seat collection, seat
	// delivery, and checkpoint failures must all leave the grain retryable.
	tests := []struct {
		name      string
		configure func(*fakeBillingReporterDB, *recordingBillingBiller, *fakeBillingMetricsClient)
		wantError string
	}{
		{
			name: "metrics fetch failure",
			configure: func(_ *fakeBillingReporterDB, _ *recordingBillingBiller, client *fakeBillingMetricsClient) {
				client.fetchErr = errors.New("metrics unavailable")
			},
			wantError: "failed to get usage metrics",
		},
		{
			name: "provider delivery failure",
			configure: func(_ *fakeBillingReporterDB, biller *recordingBillingBiller, _ *fakeBillingMetricsClient) {
				biller.failRuntime = errors.New("provider unavailable")
			},
			wantError: "failed to report usage",
		},
		{
			name: "seat collection failure",
			configure: func(db *fakeBillingReporterDB, _ *recordingBillingBiller, _ *fakeBillingMetricsClient) {
				db.seatErr = errors.New("member query failed")
			},
			wantError: "failed to report admin usage metrics",
		},
		{
			name: "seat delivery failure",
			configure: func(_ *fakeBillingReporterDB, biller *recordingBillingBiller, _ *fakeBillingMetricsClient) {
				biller.failSeats = errors.New("seat delivery failed")
			},
			wantError: "failed to report admin usage metrics",
		},
		{
			name: "checkpoint persistence failure",
			configure: func(db *fakeBillingReporterDB, _ *recordingBillingBiller, _ *fakeBillingMetricsClient) {
				db.updateFailures = 1
			},
			wantError: "failed to update last usage reporting time",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			worker, db, biller, client := newBillingReporterFixture(makeBillingMetricRows(1))
			tt.configure(db, biller, client)

			err := worker.Work(t.Context(), nil)
			require.ErrorContains(t, err, tt.wantError)
			require.True(t, db.checkpoint.IsZero())
		})
	}
}

func TestBillingReporterRetryAfterCheckpointFailureReplaysStableUsage(t *testing.T) {
	// Provider delivery can succeed immediately before checkpoint persistence
	// fails. A retry must replay the same logical records so provider idempotency can deduplicate them.
	worker, db, biller, _ := newBillingReporterFixture(makeBillingMetricRows(1))
	db.updateFailures = 1

	require.ErrorContains(t, worker.Work(t.Context(), nil), "failed to update last usage reporting time")
	require.NoError(t, worker.Work(t.Context(), nil))
	require.Len(t, biller.delivered, 4) // runtime + seats on each attempt
	require.Equal(t, biller.delivered[0], biller.delivered[2])
	require.Equal(t, biller.delivered[1], biller.delivered[3])
	require.False(t, db.checkpoint.IsZero())
}

type billingMetricsCall struct {
	startTime           time.Time
	endTime             time.Time
	afterTime           time.Time
	afterOrgID          string
	afterProjectID      string
	afterInstanceID     string
	afterBillingService string
	afterEventName      string
	limit               int
}

type fakeBillingMetricsClient struct {
	rows     []*metrics.Usage
	calls    []billingMetricsCall
	fetchErr error
}

func (c *fakeBillingMetricsClient) GetUsageMetrics(_ context.Context, startTime, endTime, afterTime time.Time, afterOrgID, afterProjectID, afterInstanceID, afterBillingService, afterEventName, _ string, limit int) ([]*metrics.Usage, error) {
	c.calls = append(c.calls, billingMetricsCall{
		startTime: startTime, endTime: endTime, afterTime: afterTime,
		afterOrgID: afterOrgID, afterProjectID: afterProjectID, afterInstanceID: afterInstanceID,
		afterBillingService: afterBillingService, afterEventName: afterEventName, limit: limit,
	})
	if c.fetchErr != nil {
		return nil, c.fetchErr
	}
	offset := 0
	if !afterTime.IsZero() {
		offset = len(c.rows)
		for i, row := range c.rows {
			if row.StartTime.Equal(afterTime) && row.OrgID == afterOrgID && row.ProjectID == afterProjectID && row.InstanceID == afterInstanceID && row.BillingService == afterBillingService && row.EventName == afterEventName {
				offset = i + 1
				break
			}
		}
	}
	if offset >= len(c.rows) {
		return nil, nil
	}
	end := min(offset+limit, len(c.rows))
	return c.rows[offset:end], nil
}

type fakeBillingReporterDB struct {
	checkpoint      time.Time
	updates         []time.Time
	updateFailures  int
	seatCount       int
	seatErr         error
	organizationIDs []string
}

func (d *fakeBillingReporterDB) FindBillingUsageReportedOn(context.Context) (time.Time, error) {
	return d.checkpoint, nil
}

func (d *fakeBillingReporterDB) UpdateBillingUsageReportedOn(_ context.Context, checkpoint time.Time) error {
	d.updates = append(d.updates, checkpoint)
	if d.updateFailures > 0 {
		d.updateFailures--
		return errors.New("checkpoint unavailable")
	}
	d.checkpoint = checkpoint
	return nil
}

func (d *fakeBillingReporterDB) CountOrganizationMemberUsers(context.Context, string, string, string, bool) (int, error) {
	return d.seatCount, d.seatErr
}

func (d *fakeBillingReporterDB) FindOrganizationIDsWithBilling(context.Context) ([]string, error) {
	return d.organizationIDs, nil
}

func (d *fakeBillingReporterDB) CountBillingProjectsForOrganization(context.Context, string, time.Time) (int, error) {
	return 0, nil
}

type recordingBillingBiller struct {
	billing.Biller
	delivered   [][]*billing.Usage
	failRuntime error
	failSeats   error
}

func (b *recordingBillingBiller) GetReportingGranularity() billing.UsageReportingGranularity {
	return billing.UsageReportingGranularityHour
}

func (b *recordingBillingBiller) ReportUsage(_ context.Context, usage []*billing.Usage) error {
	if len(usage) > 0 && usage[0].MetricName == "seats" && b.failSeats != nil {
		return b.failSeats
	}
	if len(usage) > 0 && usage[0].MetricName != "seats" && b.failRuntime != nil {
		return b.failRuntime
	}
	cloned := make([]*billing.Usage, len(usage))
	for i, item := range usage {
		copy := *item
		copy.Metadata = make(map[string]interface{}, len(item.Metadata))
		for key, value := range item.Metadata {
			copy.Metadata[key] = value
		}
		cloned[i] = &copy
	}
	b.delivered = append(b.delivered, cloned)
	return nil
}

func newBillingReporterFixture(rows []*metrics.Usage) (*BillingReporterWorker, *fakeBillingReporterDB, *recordingBillingBiller, *fakeBillingMetricsClient) {
	db := &fakeBillingReporterDB{seatCount: 3}
	biller := &recordingBillingBiller{}
	client := &fakeBillingMetricsClient{rows: rows}
	worker := &BillingReporterWorker{
		admin:    &admin.Service{Biller: biller},
		logger:   zap.NewNop(),
		database: db,
		openMetricsProject: func(context.Context) (billingUsageMetricsClient, bool, error) {
			return client, true, nil
		},
		now: func() time.Time {
			return time.Date(2026, time.January, 2, 5, 55, 0, 0, time.UTC)
		},
	}
	return worker, db, biller, client
}

func makeBillingMetricRows(count int) []*metrics.Usage {
	rows := make([]*metrics.Usage, count)
	billingCustomerID := "customer-1"
	for i := range rows {
		rows[i] = &metrics.Usage{
			OrgID:             "org-1",
			ProjectID:         fmt.Sprintf("project-%05d", i),
			ProjectName:       "Project",
			BillingCustomerID: &billingCustomerID,
			StartTime:         time.Date(2026, time.January, 2, 3, 0, 0, 0, time.UTC),
			EndTime:           time.Date(2026, time.January, 2, 4, 0, 0, 0, time.UTC),
			EventName:         "query",
			MaxValue:          1,
			SumValue:          1,
			BillingService:    "runtime",
			InstanceID:        fmt.Sprintf("instance-%05d", i),
		}
	}
	return rows
}

func countReportedMetric(batches [][]*billing.Usage, metric string) int {
	count := 0
	for _, batch := range batches {
		for _, usage := range batch {
			if usage.MetricName == metric {
				count++
			}
		}
	}
	return count
}

func findBillingUsageBatch(t *testing.T, batches [][]*billing.Usage, metric string) []*billing.Usage {
	t.Helper()
	for _, batch := range batches {
		if len(batch) > 0 && batch[0].MetricName == metric {
			return batch
		}
	}
	require.FailNow(t, "usage batch not found", metric)
	return nil
}
