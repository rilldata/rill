package slack

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestAIAlertFail(t *testing.T) {
	var text string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var msg struct {
			Text string `json:"text"`
		}
		require.NoError(t, json.Unmarshal(body, &msg))
		text = msg.Text
	}))
	defer srv.Close()

	n, err := newNotifier("", EncodeProps(nil, nil, []string{srv.URL}))
	require.NoError(t, err)
	err = n.SendAlertStatus(&drivers.AlertStatus{
		DisplayName:   "Revenue drop",
		ExecutionTime: time.Date(2026, 10, 27, 7, 0, 0, 0, time.UTC),
		Status:        runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL,
		FailRow:       map[string]any{"reason": "Revenue fell 40% against the previous four weeks.", "ai_session_id": "s1"},
		Reason:        "Revenue fell 40% against the previous four weeks.",
		OpenLink:      "https://example.com/alerts/a1/open?session_id=s1",
		EditLink:      "https://example.com/alerts/a1",
	})
	require.NoError(t, err)

	// The message shows the reason instead of the fail row's fields
	require.Contains(t, text, "Revenue fell 40% against the previous four weeks.")
	require.Contains(t, text, "https://example.com/alerts/a1/open?session_id=s1")
	require.NotContains(t, text, "first row")
	require.NotContains(t, text, "ai_session_id")
}
