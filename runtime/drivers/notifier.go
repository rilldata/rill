package drivers

import (
	"fmt"
	"time"

	"github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Notifier sends notifications.
type Notifier interface {
	SendAlertStatus(s *AlertStatus, spec NotifierSpec) error
	SendScheduledReport(s *ScheduledReport, spec NotifierSpec) error
}

type NotifierSpec = interface{}

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

func NotifierConnectorName(spec NotifierSpec) (string, error) {
	switch spec := spec.(type) {
	case *runtimev1.NotifierSpec_Email:
		return "email", nil
	case *runtimev1.NotifierSpec_Slack:
		return "slack", nil
	default:
		return "", fmt.Errorf("unknown notifier spec type: %T", spec)
	}
}
