package activity

import (
	"encoding/json"
	"time"
)

// Event is a telemetry event. It consists of a few required fields that are common to all events and a payload of type-specific data.
// All the common fields are prefixed with "event_" to avoid conflicts with the payload data.
type Event struct {
	EventID   string
	EventTime time.Time
	EventType string
	EventName string
	Data      map[string]any
}

func (e Event) MarshalJSON() ([]byte, error) {
	// Add the common fields to the map.
	if e.Data == nil {
		e.Data = make(map[string]any)
	}
	e.Data["event_id"] = e.EventID
	e.Data["event_time"] = e.EventTime.UTC().Format(time.RFC3339Nano)
	e.Data["event_type"] = e.EventType
	e.Data["event_name"] = e.EventName

	// Then serialize it.
	return json.Marshal(e.Data)
}

// Constants for common event types.
const (
	EventTypeLog        = "log"
	EventTypeMetric     = "metric"
	EventTypeBehavioral = "behavioral"
)

// Constants for common event attribute keys.
const (
	AttrKeyServiceName    = "service_name"
	AttrKeyServiceVersion = "service_version"
	AttrKeyServiceCommit  = "service_commit"
	AttrKeyIsDev          = "is_dev"
	AttrKeyInstallID      = "install_id"
	AttrKeyUserID         = "user_id"
)

// Constants for event names of type EventTypeBehavioral.
// Note: This list is not exhaustive. Proxied events may contain other names.
const (
	BehavioralEventInstallSuccess         = "install-success"
	BehavioralEventAppStart               = "app-start"
	BehavioralEventAppStop                = "app-stop"
	BehavioralEventLoginStart             = "login-start"
	BehavioralEventLoginSuccess           = "login-success"
	BehavioralEventDeployStart            = "deploy-start"
	BehavioralEventDeploySuccess          = "deploy-success"
	BehavioralEventGithubConnectedStart   = "ghconnected-start"
	BehavioralEventGithubConnectedSuccess = "ghconnected-success"
	BehavioralEventDataAccessStart        = "dataaccess-start"
	BehavioralEventDataAccessSuccess      = "dataaccess-success"
	BehavioralEventAPIQueryStart          = "api-query-start"
	BehavioralEventAPIQuerySuccess        = "api-query-success"
	BehavioralEventCanvasResolveStart     = "canvas-resolve-start"
	BehavioralEventCanvasResolveSuccess   = "canvas-resolve-success"
)

// AI-related Behavioral Events
// Note: This lists format will remain snake_case for consistency with the existing metrics.
const (
	BehavioralEventAIGeneratedRenderer    = "ai_generated_renderer"
	BehavioralEventAIGeneratedMetricsView = "ai_generated_metrics_view_yaml"
	BehavioralEventAIGeneratedResolver    = "ai_generated_resolver"
)
