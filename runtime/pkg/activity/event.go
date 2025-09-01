package activity

import (
	"encoding/json"
	"time"
)

const maxSize = 32 * 1024 // 32 KB

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

	b, err := json.Marshal(e.Data)
	if err != nil {
		return nil, err
	}
	if len(b) <= maxSize {
		return b, nil
	}

	// Truncate if too large
	truncated := map[string]any{
		"event_id":   e.EventID,
		"event_time": e.EventTime.UTC().Format(time.RFC3339Nano),
		"event_type": e.EventType,
		"event_name": e.EventName,
		"truncated":  true,
		"reason":     "event data exceeded 32KB and was truncated",
	}
	return json.Marshal(truncated)
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
	BehavioralEventLoginStart             = "login-start"
	BehavioralEventLoginSuccess           = "login-success"
	BehavioralEventDeployStart            = "deploy-start"
	BehavioralEventDeploySuccess          = "deploy-success"
	BehavioralEventGithubConnectedStart   = "ghconnected-start"
	BehavioralEventGithubConnectedSuccess = "ghconnected-success"
	BehavioralEventDataAccessStart        = "dataaccess-start"
	BehavioralEventDataAccessSuccess      = "dataaccess-success"
)
