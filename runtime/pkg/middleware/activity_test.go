package middleware

import (
	"context"
	"testing"
	"time"

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

func TestRequestMetric(t *testing.T) {
	tests := []struct {
		name       string
		claims     *runtime.SecurityClaims
		wantEmbed  bool
		wantUserID string // "" = expect no user_id attribute
	}{
		{
			name:       "external embedded user",
			claims:     &runtime.SecurityClaims{UserID: "ext_abc123", UserAttributes: map[string]any{"embed": true}},
			wantEmbed:  true,
			wantUserID: "ext_abc123",
		},
		{
			name:      "anonymous embedded user gets a synthesized id",
			claims:    &runtime.SecurityClaims{UserID: "", UserAttributes: map[string]any{"embed": true, "team": "acme"}},
			wantEmbed: true,
			// user_id is the attribute hash; asserted as non-empty below.
		},
		{
			name:       "regular user",
			claims:     &runtime.SecurityClaims{UserID: "user123"},
			wantEmbed:  false,
			wantUserID: "user123",
		},
		{
			name:      "anonymous non-embed request has no user_id",
			claims:    &runtime.SecurityClaims{UserID: ""},
			wantEmbed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sink := &captureSink{}
			client := activity.NewClient(sink, zap.NewNop())

			ctx := setRequestActivityAttributes(context.Background(), tt.claims, "/svc/Method")
			ctx = runtime.WithRequestSource(ctx, runtime.RequestSourceUI)
			emitRequestMetric(ctx, client, tt.claims, time.Now(), "OK")

			require.Len(t, sink.events, 1)
			e := sink.events[0]
			require.Equal(t, "request_time_ms", e.EventName)
			require.Equal(t, "ui", e.Data["source"])
			require.Equal(t, tt.wantEmbed, e.Data["embed"])

			if tt.wantEmbed && tt.wantUserID == "" {
				// anonymous embedded user: a non-empty synthesized id
				require.NotEmpty(t, e.Data["user_id"])
			} else if tt.wantUserID != "" {
				require.Equal(t, tt.wantUserID, e.Data["user_id"])
			} else {
				require.Nil(t, e.Data["user_id"])
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
