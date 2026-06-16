package middleware

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// captureSink records emitted events for assertions.
type captureSink struct {
	events []activity.Event
}

func (s *captureSink) Emit(e activity.Event) error {
	s.events = append(s.events, e)
	return nil
}

func (s *captureSink) Close() {}

// fakeRequest implements the GetInstanceId interface used to attribute usage.
type fakeRequest struct {
	instanceID string
}

func (r fakeRequest) GetInstanceId() string { return r.instanceID }

func eventNames(events []activity.Event) []string {
	var names []string
	for _, e := range events {
		names = append(names, e.EventName)
	}
	return names
}

func TestRecordEmbeddedUserAPICall(t *testing.T) {
	instanceAttrs := func(ctx context.Context, instanceID string) []attribute.KeyValue {
		return []attribute.KeyValue{attribute.String("org_id", "org1"), attribute.String("project_id", "proj1")}
	}

	tests := []struct {
		name       string
		claims     *runtime.SecurityClaims
		req        interface{}
		wantMetric string // "" means nothing emitted
	}{
		{
			name:       "external user",
			claims:     &runtime.SecurityClaims{UserID: "ext_abc123"},
			req:        fakeRequest{instanceID: "inst1"},
			wantMetric: "external_user_api_call",
		},
		{
			name:       "anonymous embedded user",
			claims:     &runtime.SecurityClaims{UserID: "", UserAttributes: map[string]any{"embed": true, "team": "acme"}},
			req:        fakeRequest{instanceID: "inst1"},
			wantMetric: "external_anonymous_user_api_call",
		},
		{
			name:       "owner previewing embed is not anonymous",
			claims:     &runtime.SecurityClaims{UserID: "user123", UserAttributes: map[string]any{"embed": true}},
			req:        fakeRequest{instanceID: "inst1"},
			wantMetric: "",
		},
		{
			name:       "regular dashboard user",
			claims:     &runtime.SecurityClaims{UserID: "user123", UserAttributes: map[string]any{"email": "a@b.com"}},
			req:        fakeRequest{instanceID: "inst1"},
			wantMetric: "",
		},
		{
			name:       "external user without resolvable instance is skipped",
			claims:     &runtime.SecurityClaims{UserID: "ext_abc123"},
			req:        struct{}{},
			wantMetric: "",
		},
		{
			name:       "nil claims",
			claims:     nil,
			req:        fakeRequest{instanceID: "inst1"},
			wantMetric: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sink := &captureSink{}
			client := activity.NewClient(sink, zap.NewNop())

			recordEmbeddedUserAPICall(context.Background(), client, instanceAttrs, tt.claims, tt.req)

			if tt.wantMetric == "" {
				require.Empty(t, eventNames(sink.events))
				return
			}
			require.Equal(t, []string{tt.wantMetric}, eventNames(sink.events))
			// Usage must be attributed to the org and project.
			require.Equal(t, "org1", sink.events[0].Data["org_id"])
			require.Equal(t, "proj1", sink.events[0].Data["project_id"])
			// Anonymous embedded users carry an "anon_"-prefixed user_id for distinct counting.
			if tt.wantMetric == "external_anonymous_user_api_call" {
				require.Contains(t, sink.events[0].Data["user_id"], "anon_")
			}
		})
	}
}

func TestAnonymousUserID(t *testing.T) {
	// Deterministic and independent of map iteration order.
	a := anonymousUserID(map[string]any{"team": "acme", "region": "us", "embed": true})
	b := anonymousUserID(map[string]any{"region": "us", "embed": true, "team": "acme"})
	require.Equal(t, a, b)

	// Different attributes produce a different identifier.
	c := anonymousUserID(map[string]any{"team": "globex", "region": "us", "embed": true})
	require.NotEqual(t, a, c)
}
