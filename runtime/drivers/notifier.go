package drivers

import (
	"html/template"
	"time"

	"github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Notifier sends notifications.
type Notifier interface {
	SendAlertStatus(s *AlertStatus, r RecipientOpts) error
}

type RecipientOpts interface{}

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

type AlertStatusData struct {
	Subject             string
	Title               string
	ExecutionTimeString string // Will be inferred from ExecutionTime
	IsPass              bool
	IsRecover           bool
	IsError             bool
	ErrorMessage        string
	OpenLink            template.URL
	EditLink            template.URL
}

type AlertFailData struct {
	Subject             string
	Title               string
	ExecutionTimeString string // Will be inferred from ExecutionTime
	FailRow             map[string]any
	OpenLink            template.URL
	EditLink            template.URL
}
