package slack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestSendAlertFail(t *testing.T) {
	var texts []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Text string `json:"text"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		texts = append(texts, payload.Text)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n, err := newNotifier("", EncodeProps(nil, nil, []string{srv.URL}))
	require.NoError(t, err)

	status := &drivers.AlertStatus{
		DisplayName:   "Orders by country",
		ExecutionTime: time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC),
		Status:        runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL,
		FailColumns:   []string{"country", "Orders", "Orders (Δ%)"},
		OpenLink:      "https://example.com/open",
		EditLink:      "https://example.com/edit",
	}

	// A single row lists its fields
	status.FailRows = [][]string{{"US", "12.3k", "-8%"}}
	require.NoError(t, n.SendAlertStatus(status))
	require.Len(t, texts, 1)
	require.Contains(t, texts[0], "The first row that matched your alert criteria is:\n\n• *country*: US\n• *Orders*: 12.3k\n• *Orders (Δ%)*: -8%\n")
	require.NotContains(t, texts[0], "More rows matched")

	// Several rows are listed one per line, with values Slack would interpret as mentions escaped
	status.FailRows = [][]string{{"US", "12.3k", "-8%"}, {"<!channel>", "4.5k", "-12%"}}
	require.NoError(t, n.SendAlertStatus(status))
	require.Len(t, texts, 2)
	require.Contains(t, texts[1], "The rows that matched your alert criteria are:\n\n• *US*: Orders 12.3k, Orders (Δ%) -8%\n• *&lt;!channel&gt;*: Orders 4.5k, Orders (Δ%) -12%\n")
	require.NotContains(t, texts[1], "<!channel>")
	require.NotContains(t, texts[1], "More rows matched")

	// More rows matched than are listed
	status.FailRowsTruncated = true
	require.NoError(t, n.SendAlertStatus(status))
	require.Len(t, texts, 3)
	require.Contains(t, texts[2], "The first 2 rows that matched your alert criteria are:")
	require.Contains(t, texts[2], "More rows matched your alert criteria.")
}

func TestSlackText(t *testing.T) {
	require.Equal(t, "a &amp; b &lt;c&gt;", slackText("a & b <c>"))

	long := strings.Repeat("x", slackTextMaxLen+10)
	truncated := slackText(long)
	require.Equal(t, slackTextMaxLen, len([]rune(truncated)))
	require.True(t, strings.HasSuffix(truncated, "…"))
}
