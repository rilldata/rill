package activity_test

// test MarshalJSON

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/assert"
)

func TestEventMarshalJSON_Truncate(t *testing.T) {
	t.Run("truncated event", func(t *testing.T) {
		// Create a large event data to exceed the max size
		data := make(map[string]any)
		for i := 0; i < 100000; i++ {
			data[fmt.Sprintf("key-%d", i)] = "x"
		}

		event := activity.Event{
			EventID:   "12345",
			EventTime: time.Date(2023, 10, 1, 12, 0, 0, 0, time.UTC),
			EventType: activity.EventTypeLog,
			EventName: "test_event",
			Data:      data,
		}

		dataBytes, err := json.Marshal(event)
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(dataBytes, &result)
		assert.NoError(t, err)

		assert.Equal(t, "12345", result["event_id"])
		assert.Equal(t, "2023-10-01T12:00:00Z", result["event_time"])
		assert.Equal(t, activity.EventTypeLog, result["event_type"])
		assert.Equal(t, "test_event", result["event_name"])
		assert.True(t, result["truncated"].(bool))
		assert.Equal(t, "event data exceeded 32KB and was truncated", result["reason"])

		// {"event_id":"12345","event_name":"test_event","event_time":"2023-10-01T12:00:00Z","event_type":"log","reason":"event data exceeded 1MB and was truncated","truncated":true}
	})

	t.Run("non-truncated event", func(t *testing.T) {
		// Create a small event data to stay under the max size
		data := map[string]any{
			"foo": "bar",
			"baz": 123,
		}

		event := activity.Event{
			EventID:   "67890",
			EventTime: time.Date(2023, 11, 1, 12, 0, 0, 0, time.UTC),
			EventType: activity.EventTypeMetric,
			EventName: "small_event",
			Data:      data,
		}

		dataBytes, err := json.Marshal(event)
		assert.NoError(t, err)

		var result map[string]any
		err = json.Unmarshal(dataBytes, &result)
		assert.NoError(t, err)

		assert.Equal(t, "67890", result["event_id"])
		assert.Equal(t, "2023-11-01T12:00:00Z", result["event_time"])
		assert.Equal(t, activity.EventTypeMetric, result["event_type"])
		assert.Equal(t, "small_event", result["event_name"])
		assert.Equal(t, "bar", result["foo"])
		assert.EqualValues(t, 123, result["baz"])
		_, truncated := result["truncated"]
		assert.False(t, truncated)
	})
}
