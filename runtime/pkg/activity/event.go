package activity

import (
	"encoding/json"
	"time"
)

// Event is a telemetry event. It consists of a few required common fields and a payload of event-specific data.
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
	return json.Marshal(e.Data)
}

// Constants for common event types.
const (
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
