package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func newTestNotifier(t *testing.T, config *configProperties, urls []string) *notifier {
	t.Helper()
	n, err := newNotifier(config, EncodeProps(urls))
	require.NoError(t, err)
	// Keep tests fast.
	n.retryWaitMin = time.Millisecond
	n.retryWaitMax = 5 * time.Millisecond
	n.requestTimeout = 5 * time.Second
	return n
}

func testAlertStatus() *drivers.AlertStatus {
	return &drivers.AlertStatus{
		DisplayName:   "Test Alert",
		ExecutionTime: time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC),
		Status:        runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL,
		FailRow:       map[string]any{"region": "south", "sales": 0.0},
		OpenLink:      "https://example.com/open",
		EditLink:      "https://example.com/edit",
	}
}

// TestSignOfficialVector verifies the signature scheme against the example published in the
// Standard Webhooks specification / Svix documentation.
func TestSignOfficialVector(t *testing.T) {
	secret := "whsec_MfKQ9r8GKYqrTwjUPD8ILPZIo2LaLaSw"
	id := "msg_p5jXN8AQM9LWM0D4loKWxJek"
	ts := time.Unix(1614265330, 0)
	body := []byte(`{"test": 2432232314}`)

	got, err := sign(secret, id, ts, body)
	require.NoError(t, err)
	require.Equal(t, "v1,g0hM9SsE+OTPJTGt/tmIKtSyZlE3uFJELVlNIOLJ1OE=", got)
}

func TestSigningKey(t *testing.T) {
	// A secret without the whsec_ prefix is used as raw key bytes.
	key, err := signingKey("raw-secret")
	require.NoError(t, err)
	require.Equal(t, []byte("raw-secret"), key)

	// Invalid base64 after the prefix errors without leaking the secret value.
	_, err = signingKey("whsec_!!!not-base64!!!")
	require.Error(t, err)
	require.NotContains(t, err.Error(), "not-base64")
}

func TestSendAlertStatusPayloadAndSignature(t *testing.T) {
	secret := "whsec_MfKQ9r8GKYqrTwjUPD8ILPZIo2LaLaSw"

	var gotBody []byte
	var gotHeaders http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		gotHeaders = r.Header.Clone()
		require.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := newTestNotifier(t, &configProperties{
		SigningSecret: secret,
		Headers:       map[string]string{"Authorization": "Bearer tok123"},
	}, []string{srv.URL})

	require.NoError(t, n.SendAlertStatus(testAlertStatus()))

	// Envelope
	var p struct {
		ID        string    `json:"id"`
		Type      string    `json:"type"`
		Version   int       `json:"version"`
		Timestamp time.Time `json:"timestamp"`
		Data      struct {
			DisplayName string         `json:"display_name"`
			Status      string         `json:"status"`
			IsRecover   bool           `json:"is_recover"`
			FailRow     map[string]any `json:"fail_row"`
			OpenLink    string         `json:"open_link"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(gotBody, &p))
	require.NotEmpty(t, p.ID)
	require.Equal(t, "alert.status", p.Type)
	require.Equal(t, 1, p.Version)
	require.Equal(t, "Test Alert", p.Data.DisplayName)
	require.Equal(t, "FAIL", p.Data.Status)
	require.Equal(t, map[string]any{"region": "south", "sales": 0.0}, p.Data.FailRow)
	require.Equal(t, "https://example.com/open", p.Data.OpenLink)

	// Static headers
	require.Equal(t, "application/json", gotHeaders.Get("Content-Type"))
	require.Equal(t, "Bearer tok123", gotHeaders.Get("Authorization"))

	// Standard Webhooks headers: id matches the envelope, timestamp matches the payload,
	// and the signature verifies against the raw body.
	require.Equal(t, p.ID, gotHeaders.Get("webhook-id"))
	tsHeader := gotHeaders.Get("webhook-timestamp")
	require.NotEmpty(t, tsHeader)

	key, err := signingKey(secret)
	require.NoError(t, err)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(p.ID + "." + tsHeader + "."))
	mac.Write(gotBody)
	want := "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
	require.Equal(t, want, gotHeaders.Get("webhook-signature"))
}

func TestSendUnsignedWithoutSecret(t *testing.T) {
	var gotHeaders http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := newTestNotifier(t, &configProperties{}, []string{srv.URL})
	require.NoError(t, n.SendAlertStatus(testAlertStatus()))

	require.Empty(t, gotHeaders.Get("webhook-id"))
	require.Empty(t, gotHeaders.Get("webhook-timestamp"))
	require.Empty(t, gotHeaders.Get("webhook-signature"))
}

func TestSendScheduledReport(t *testing.T) {
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := newTestNotifier(t, &configProperties{}, []string{srv.URL})
	require.NoError(t, n.SendScheduledReport(&drivers.ScheduledReport{
		DisplayName:    "Test Report",
		ReportTime:     time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC),
		DownloadFormat: "csv",
		OpenLink:       "https://example.com/open",
		DownloadLink:   "https://example.com/download",
	}))

	var p struct {
		Type string `json:"type"`
		Data struct {
			DisplayName    string `json:"display_name"`
			DownloadFormat string `json:"download_format"`
			DownloadLink   string `json:"download_link"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(gotBody, &p))
	require.Equal(t, "report.scheduled", p.Type)
	require.Equal(t, "Test Report", p.Data.DisplayName)
	require.Equal(t, "csv", p.Data.DownloadFormat)
	require.Equal(t, "https://example.com/download", p.Data.DownloadLink)
}

func TestRetriesOn5xxThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := newTestNotifier(t, &configProperties{}, []string{srv.URL})
	require.NoError(t, n.SendAlertStatus(testAlertStatus()))
	require.Equal(t, int32(3), calls.Load())
}

func TestNoRetryOn4xx(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	n := newTestNotifier(t, &configProperties{}, []string{srv.URL})
	err := n.SendAlertStatus(testAlertStatus())
	require.Error(t, err)
	// The error must be self-sufficient: it is the only debugging surface in Rill.
	require.Contains(t, err.Error(), srv.URL)
	require.Contains(t, err.Error(), "unexpected status 400")
	require.Equal(t, int32(1), calls.Load())
}

func TestAllURLsAttemptedDespiteFailure(t *testing.T) {
	var failingCalls, okCalls atomic.Int32
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failingCalls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failing.Close()
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okCalls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ok.Close()

	n := newTestNotifier(t, &configProperties{}, []string{failing.URL, ok.URL})
	n.retryMax = 0 // single attempt to keep the test fast

	err := n.SendAlertStatus(testAlertStatus())
	require.Error(t, err)
	require.Contains(t, err.Error(), failing.URL)
	require.NotContains(t, err.Error(), ok.URL)
	require.Equal(t, int32(1), failingCalls.Load())
	require.Equal(t, int32(1), okCalls.Load())
}

func TestDuplicateURLsSentOnce(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := newTestNotifier(t, &configProperties{}, []string{srv.URL, srv.URL})
	require.NoError(t, n.SendAlertStatus(testAlertStatus()))
	require.Equal(t, int32(1), calls.Load())
}

func TestInvalidSecretFailsFast(t *testing.T) {
	_, err := newNotifier(&configProperties{SigningSecret: "whsec_!!!"}, EncodeProps([]string{"https://example.com"}))
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "invalid signing secret"))
}
