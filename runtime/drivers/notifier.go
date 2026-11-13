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
	DisplayName       string
	ExecutionTime     time.Time
	Status            runtimev1.AssertionStatus
	IsRecover         bool
	FailColumns       []string   // Column names of FailRows, in the order of the alert query's result
	FailRows          [][]string // Display values of the rows that matched the alert criteria; it has at most the alert's rows limit
	FailRowsTruncated bool       // True if more rows matched the alert criteria than FailRows contains
	ExecutionError    string
	OpenLink          string
	EditLink          string
	UnsubscribeLink   string
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
