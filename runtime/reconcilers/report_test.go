package reconcilers

import (
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestNextRefreshTimeBoundaries(t *testing.T) {
	// Resolve disabled, ticker, cron, timezone, and invalid schedules at a fixed instant.
	now := time.Date(2026, time.January, 2, 10, 30, 0, 0, time.UTC)

	// Schedules are persisted as the next instant strictly after now; this table
	// covers disabled, ticker, cron, timezone, and precedence behavior.
	tests := []struct {
		name     string
		schedule *runtimev1.Schedule
		want     time.Time
		wantErr  string
	}{
		{name: "nil schedule"},
		{name: "disabled", schedule: &runtimev1.Schedule{Disable: true}},
		{name: "ticker", schedule: &runtimev1.Schedule{TickerSeconds: 30}, want: now.Add(30 * time.Second)},
		{name: "cron", schedule: &runtimev1.Schedule{Cron: "0 * * * *"}, want: time.Date(2026, time.January, 2, 11, 0, 0, 0, time.UTC)},
		{name: "timezone cron", schedule: &runtimev1.Schedule{Cron: "0 6 * * *", TimeZone: "America/Los_Angeles"}, want: time.Date(2026, time.January, 2, 14, 0, 0, 0, time.UTC)},
		{name: "earliest of ticker and cron", schedule: &runtimev1.Schedule{TickerSeconds: 10, Cron: "0 * * * *"}, want: now.Add(10 * time.Second)},
		{name: "invalid cron", schedule: &runtimev1.Schedule{Cron: "not a cron"}, wantErr: "failed to parse cron schedule"},
		{name: "invalid timezone", schedule: &runtimev1.Schedule{Cron: "0 * * * *", TimeZone: "Mars/Olympus"}, wantErr: "failed to parse cron schedule"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nextRefreshTime(now, tt.schedule)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCalculateReportExecutionTimesWatermarkAndValidation(t *testing.T) {
	// Watermark invariants prevent duplicate delivery, while malformed interval configuration must fail explicitly.
	watermark := time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC)

	t.Run("unchanged watermark is skipped", func(t *testing.T) {
		// Re-running an inherited watermark would duplicate every downstream delivery.
		_, err := calculateReportExecutionTimes(&runtimev1.Report{Spec: &runtimev1.ReportSpec{}}, watermark, watermark)
		require.ErrorContains(t, err, "watermark is unchanged")
	})

	t.Run("no interval uses exact watermark", func(t *testing.T) {
		// Reports without interval expansion must preserve the source watermark exactly.
		got, err := calculateReportExecutionTimes(&runtimev1.Report{Spec: &runtimev1.ReportSpec{}}, watermark, time.Time{})
		require.NoError(t, err)
		require.Equal(t, []time.Time{watermark}, got)
	})

	for _, tt := range []struct {
		name     string
		interval string
		timezone string
		wantErr  string
	}{
		{name: "malformed duration", interval: "definitely-not-ISO-8601", wantErr: "failed to parse interval duration"},
		{name: "nonstandard duration", interval: "inf", wantErr: "is not a standard ISO 8601 duration"},
		{name: "invalid timezone", interval: "P1D", timezone: "Mars/Olympus", wantErr: "failed to load time zone"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Parser validation is not a security boundary; reconciler inputs must still fail safely.
			report := &runtimev1.Report{Spec: &runtimev1.ReportSpec{
				IntervalsIsoDuration: tt.interval,
				RefreshSchedule:      &runtimev1.Schedule{TimeZone: tt.timezone},
			}}
			_, err := calculateReportExecutionTimes(report, watermark, time.Time{})
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestCalculateReportExecutionTimesFloorCeilAndLimit(t *testing.T) {
	// Compare closed and open interval rounding and pin the exact execution-count limit.
	watermark := time.Date(2026, time.January, 10, 12, 34, 0, 0, time.UTC)
	previous := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	t.Run("closed intervals floor and respect the configured count", func(t *testing.T) {
		// The limit counts executions, not loop iterations; returning limit+1 can
		// unexpectedly fan out costly reports and notifications.
		report := &runtimev1.Report{Spec: &runtimev1.ReportSpec{
			IntervalsIsoDuration: "P1D",
			IntervalsLimit:       2,
		}}
		got, err := calculateReportExecutionTimes(report, watermark, previous)
		require.NoError(t, err)
		require.Equal(t, []time.Time{
			time.Date(2026, time.January, 9, 0, 0, 0, 0, time.UTC),
			time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC),
		}, got)
	})

	t.Run("unclosed intervals ceil", func(t *testing.T) {
		// Opting into an open interval rounds forward while still returning times chronologically.
		report := &runtimev1.Report{Spec: &runtimev1.ReportSpec{
			IntervalsIsoDuration:   "P1D",
			IntervalsLimit:         2,
			IntervalsCheckUnclosed: true,
		}}
		got, err := calculateReportExecutionTimes(report, watermark, previous)
		require.NoError(t, err)
		require.Equal(t, []time.Time{
			time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC),
			time.Date(2026, time.January, 11, 0, 0, 0, 0, time.UTC),
		}, got)
	})

	t.Run("closed interval not beyond previous watermark is skipped", func(t *testing.T) {
		// A partial interval at the same boundary must not be delivered twice.
		report := &runtimev1.Report{Spec: &runtimev1.ReportSpec{IntervalsIsoDuration: "P1D"}}
		_, err := calculateReportExecutionTimes(report, watermark, time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC))
		require.ErrorContains(t, err, "watermark has not advanced by a full interval")
	})
}

func TestCalculateReportExecutionTimesDSTBoundary(t *testing.T) {
	// Local-midnight intervals span 23 hours across the New York spring-forward
	// transition; fixed 24-hour arithmetic would select the wrong report dates.
	report := &runtimev1.Report{Spec: &runtimev1.ReportSpec{
		IntervalsIsoDuration: "P1D",
		IntervalsLimit:       2,
		RefreshSchedule:      &runtimev1.Schedule{TimeZone: "America/New_York"},
	}}
	watermark := time.Date(2026, time.March, 9, 5, 30, 0, 0, time.UTC)
	previous := time.Date(2026, time.March, 7, 5, 0, 0, 0, time.UTC)
	got, err := calculateReportExecutionTimes(report, watermark, previous)
	require.NoError(t, err)
	require.Len(t, got, 2)
	location, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	require.True(t, time.Date(2026, time.March, 8, 0, 0, 0, 0, location).Equal(got[0]))
	require.True(t, time.Date(2026, time.March, 9, 0, 0, 0, 0, location).Equal(got[1]))
	require.Equal(t, 23*time.Hour, got[1].Sub(got[0]))
}

func TestLatestReportWarningsUsesNewestHistoryEntry(t *testing.T) {
	// Execution history is newest-first because popCurrentExecution inserts at
	// index zero; surfacing the tail silently shows stale warnings to users.
	report := &runtimev1.Report{State: &runtimev1.ReportState{ExecutionHistory: []*runtimev1.ReportExecution{
		{Warnings: []string{"newest warning"}},
		{Warnings: []string{"oldest warning"}},
	}}}
	require.Equal(t, []string{"newest warning"}, latestReportWarnings(report))
	require.Nil(t, latestReportWarnings(&runtimev1.Report{}))
}
