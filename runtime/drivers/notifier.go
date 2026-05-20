package drivers

import (
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Notifier sends notifications.
type Notifier interface {
	SendAlertStatus(s *AlertStatus) error
	SendScheduledReport(s *ScheduledReport) error
}

type AlertStatus struct {
	// TODO: Remove ToEmail, ToName once email notifier is created
	ToEmail     string
	ToName      string
	DisplayName string
	ExecutionTime time.Time
	Status      runtimev1.AssertionStatus
	IsRecover   bool
	// FailRow is the first matching row. Retained for backwards compatibility with consumers that
	// haven't migrated to FailRows yet. New consumers should read FailRows.
	FailRow map[string]any
	// FailRows holds all matching rows for the alert, up to AlertSpec.NotificationRowLimit.
	FailRows []map[string]any
	// FailRowsTruncated is true when more rows matched the alert than were included in FailRows.
	// Renderers surface this as a "N+ rows matched..." indicator.
	FailRowsTruncated bool
	ExecutionError    string
	OpenLink        string
	EditLink        string
	UnsubscribeLink string
}

type ScheduledReport struct {
	DisplayName     string
	ReportTime      time.Time
	DownloadFormat  string
	OpenLink        string
	DownloadLink    string
	UnsubscribeLink string
	Summary         string
}
