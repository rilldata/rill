package activity

import (
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// Event is a telemetry event.
type Event interface {
	Marshal() ([]byte, error)
}

// MetricEvent is a telemetry event representing a system metric.
type MetricEvent struct {
	Time  time.Time
	Name  string
	Value float64
	Attrs []attribute.KeyValue
}

var _ Event = (*MetricEvent)(nil)

func (e *MetricEvent) Marshal() ([]byte, error) {
	// Create a map to hold the flattened event structure.
	flattened := make(map[string]interface{})

	// Iterate over the dims slice and add each dim to the map.
	for _, dim := range e.Attrs {
		key := string(dim.Key)
		flattened[key] = dim.Value.AsInterface()
	}

	// Add the non-dim fields.
	flattened["time"] = e.Time.UTC().Format(time.RFC3339)
	flattened["name"] = e.Name
	flattened["value"] = e.Value

	return json.Marshal(flattened)
}

// Constants for UserEvent.Action.
// Note: This is not an enum because the list is not exhaustive. Proxied events may contain other actions.
const (
	UserActionInstallSuccess         = "install-success"
	UserActionDeployStart            = "deploy-start"
	UserActionDeploySuccess          = "deploy-success"
	UserActionGithubConnectedStart   = "ghconnected-start"
	UserActionGithubConnectedSuccess = "ghconnected-success"
	UserActionDataAccessStart        = "dataaccess-start"
	UserActionDataAccessSuccess      = "dataaccess-success"
	UserActionLoginStart             = "login-start"
	UserActionLoginSuccess           = "login-success"
	UserActionAppStart               = "app-start"
)

// UserEvent is a telemetry event representing a user action.
type UserEvent struct {
	AppName       string         `json:"app_name"`
	InstallID     string         `json:"install_id"`
	BuildID       string         `json:"build_id"`
	Version       string         `json:"version"`
	UserID        string         `json:"user_id"`
	IsDev         bool           `json:"is_dev"`
	Mode          string         `json:"mode"`
	Action        string         `json:"action"`
	Medium        string         `json:"medium"`
	Space         string         `json:"space"`
	ScreenName    string         `json:"screen_name"`
	EventDatetime int64          `json:"event_datetime"`
	EventType     string         `json:"event_type"`
	Payload       map[string]any `json:"payload"`
}

var _ Event = (*UserEvent)(nil)

func (e *UserEvent) Marshal() ([]byte, error) {
	return json.Marshal(e)
}
