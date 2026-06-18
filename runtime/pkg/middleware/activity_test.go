package middleware

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
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

func eventNames(events []activity.Event) []string {
	var names []string
	for _, e := range events {
		names = append(names, e.EventName)
	}
	return names
}

func TestRecordEmbeddedUsage(t *testing.T) {
	tests := []struct {
		name        string
		claims      *runtime.SecurityClaims
		wantEmitted bool
		wantUserID  string // expected user_id on the event (empty = don't assert exact value)
	}{
		{
			name:        "external embedded user",
			claims:      &runtime.SecurityClaims{UserID: "ext_abc123", UserAttributes: map[string]any{"embed": true}},
			wantEmitted: true,
			wantUserID:  "ext_abc123",
		},
		{
			name:        "anonymous embedded user",
			claims:      &runtime.SecurityClaims{UserID: "", UserAttributes: map[string]any{"embed": true, "team": "acme"}},
			wantEmitted: true,
		},
		{
			name:        "external user without embed attribute is skipped",
			claims:      &runtime.SecurityClaims{UserID: "ext_abc123"},
			wantEmitted: false,
		},
		{
			name:        "regular dashboard user",
			claims:      &runtime.SecurityClaims{UserID: "user123", UserAttributes: map[string]any{"email": "a@b.com"}},
			wantEmitted: false,
		},
		{
			name:        "nil claims",
			claims:      nil,
			wantEmitted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sink := &captureSink{}
			client := activity.NewClient(sink, zap.NewNop())

			recordEmbeddedUsage(context.Background(), client, tt.claims)

			if !tt.wantEmitted {
				require.Empty(t, eventNames(sink.events))
				return
			}
			require.Equal(t, []string{"embedded_user_request"}, eventNames(sink.events))
			require.NotEmpty(t, sink.events[0].Data["user_id"])
			if tt.wantUserID != "" {
				require.Equal(t, tt.wantUserID, sink.events[0].Data["user_id"])
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
