package drivers

import (
	"time"

	"github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Notifier sends notifications.
type Notifier interface {
	SendAlertStatus(s *AlertStatus) error
	SendScheduledReport(s *ScheduledReport) error
}

type AlertStatus struct {
	ToEmail        string
	ToName         string
	Title          string
	ExecutionTime  time.Time
	Status         runtimev1.AssertionStatus
	IsRecover      bool
	FailRow        map[string]any
	ExecutionError string
	OpenLink       string
	EditLink       string
}

type ScheduledReport struct {
	Title          string
	ReportTime     time.Time
	DownloadFormat string
	OpenLink       string
	DownloadLink   string
	EditLink       string
}
