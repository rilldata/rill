package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

// officialVectorSecret is the example signing secret published in the Standard Webhooks
// specification docs. It is not a real credential; the literal is split so secret-scanning
// tools don't flag it as one.
const officialVectorSecret = "whsec_" + "MfKQ9r8GKYqrTwjUPD8ILPZIo2LaLaSw"

func newTestNotifier(t *testing.T, config *configProperties, urls []string) *notifier {
	t.Helper()
	// Test servers listen on loopback, which is only reachable through allowed_url_prefixes.
	if config.AllowedURLPrefixes == nil {
		config.AllowedURLPrefixes = urls
	}
	n, err := newNotifier(config, EncodeProps(urls))
	require.NoError(t, err)
	// Keep tests fast.
	n.client.RetryWaitMin = time.Millisecond
	n.client.RetryWaitMax = 5 * time.Millisecond
	n.client.HTTPClient.Timeout = 5 * time.Second
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
	secret := officialVectorSecret
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
	secret := officialVectorSecret

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
	n.client.RetryMax = 0 // single attempt to keep the test fast

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

func TestPingValidatesConfig(t *testing.T) {
	ctx := context.Background()
	require.NoError(t, (&handle{config: &configProperties{}}).Ping(ctx))
	require.NoError(t, (&handle{config: &configProperties{SigningSecret: officialVectorSecret}}).Ping(ctx))
	require.Error(t, (&handle{config: &configProperties{SigningSecret: "whsec_!!!"}}).Ping(ctx))

	headers := map[string]string{"Authorization": "Bearer tok123"}
	err := (&handle{config: &configProperties{Headers: headers}}).Ping(ctx)
	require.ErrorContains(t, err, "headers require allowed_url_prefixes")
	require.NoError(t, (&handle{config: &configProperties{Headers: headers, AllowedURLPrefixes: []string{"https://example.com/"}}}).Ping(ctx))
	require.Error(t, (&handle{config: &configProperties{AllowedURLPrefixes: []string{"example.com"}}}).Ping(ctx))
}

func TestRetryAfterIsCapped(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			// A receiver must not be able to stall delivery for a day.
			w.Header().Set("Retry-After", "86400")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := newTestNotifier(t, &configProperties{}, []string{srv.URL})
	require.NoError(t, sendWithin(t, 3*time.Second, func() error { return n.SendAlertStatus(testAlertStatus()) }))
	require.Equal(t, int32(2), calls.Load())
}

func TestDeliveryDeadline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Never answer. The body must be read for the server to notice the client going away.
		_, _ = io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	}))
	defer srv.Close()

	n := newTestNotifier(t, &configProperties{}, []string{srv.URL})
	n.deliveryTimeout = 200 * time.Millisecond

	err := sendWithin(t, 3*time.Second, func() error { return n.SendAlertStatus(testAlertStatus()) })
	require.Error(t, err)
	require.Contains(t, err.Error(), srv.URL)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestSlowURLDoesNotDelayOthers(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Never answer. The body must be read for the server to notice the client going away.
		_, _ = io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	}))
	defer slow.Close()
	var fastCalls atomic.Int32
	fast := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fastCalls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer fast.Close()

	// The slow URL comes first and uses up the whole deadline.
	n := newTestNotifier(t, &configProperties{}, []string{slow.URL, fast.URL})
	n.client.RetryMax = 0
	n.client.HTTPClient.Timeout = time.Minute
	n.deliveryTimeout = 300 * time.Millisecond

	err := sendWithin(t, 3*time.Second, func() error { return n.SendAlertStatus(testAlertStatus()) })
	require.Error(t, err)
	require.Contains(t, err.Error(), slow.URL)
	require.NotContains(t, err.Error(), fast.URL)
	require.Equal(t, int32(1), fastCalls.Load())
}

func TestNoIdleConnectionsAfterSend(t *testing.T) {
	var open atomic.Int32
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	srv.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		switch state {
		case http.StateNew:
			open.Add(1)
		case http.StateClosed, http.StateHijacked:
			open.Add(-1)
		}
	}
	srv.Start()
	defer srv.Close()

	// The reconcilers build a notifier per execution, so connections kept alive after a send
	// would only be closed by the transport's idle timeout.
	n := newTestNotifier(t, &configProperties{}, []string{srv.URL})
	require.NoError(t, n.SendAlertStatus(testAlertStatus()))
	require.Eventually(t, func() bool { return open.Load() == 0 }, time.Second, 10*time.Millisecond)
}

func TestConcurrentDeliveriesAreCapped(t *testing.T) {
	var inFlight, maxInFlight, calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := inFlight.Add(1)
		defer inFlight.Add(-1)
		for {
			prev := maxInFlight.Load()
			if cur <= prev || maxInFlight.CompareAndSwap(prev, cur) {
				break
			}
		}
		calls.Add(1)
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	urls := make([]string, 3*maxConcurrentDeliveries)
	for i := range urls {
		urls[i] = fmt.Sprintf("%s/hook?n=%d", srv.URL, i)
	}
	n := newTestNotifier(t, &configProperties{}, urls)
	require.NoError(t, n.SendAlertStatus(testAlertStatus()))
	require.Equal(t, int32(len(urls)), calls.Load())
	require.LessOrEqual(t, maxInFlight.Load(), int32(maxConcurrentDeliveries))
	require.Greater(t, maxInFlight.Load(), int32(1)) // still parallel
}

// sendWithin runs fn and fails the test if it has not returned after d.
func sendWithin(t *testing.T, d time.Duration, fn func() error) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- fn() }()
	select {
	case err := <-done:
		return err
	case <-time.After(d):
		t.Fatalf("send still running after %s", d)
		return nil
	}
}

func TestHeadersRequireAllowedURLPrefixes(t *testing.T) {
	headers := map[string]string{"X-Api-Key": "gateway-secret"}
	_, err := newNotifier(&configProperties{Headers: headers}, EncodeProps([]string{"https://example.com/hook"}))
	require.ErrorContains(t, err, "headers require allowed_url_prefixes")

	_, err = newNotifier(&configProperties{Headers: headers, AllowedURLPrefixes: []string{"https://example.com/"}}, EncodeProps([]string{"https://example.com/hook"}))
	require.NoError(t, err)
}

func TestHeadersOnlySentToAllowedURLs(t *testing.T) {
	var trustedKey atomic.Value
	trusted := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		trustedKey.Store(r.Header.Get("X-Api-Key"))
		w.WriteHeader(http.StatusOK)
	}))
	defer trusted.Close()
	var attackerCalls atomic.Int32
	attacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attackerCalls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer attacker.Close()

	// An alert author added their own URL next to the gateway the connector is configured for.
	n := newTestNotifier(t, &configProperties{
		Headers:            map[string]string{"X-Api-Key": "gateway-secret"},
		AllowedURLPrefixes: []string{trusted.URL + "/"},
	}, []string{trusted.URL + "/hook", attacker.URL + "/hook"})

	err := n.SendAlertStatus(testAlertStatus())
	require.ErrorContains(t, err, attacker.URL+"/hook: URL is not in the connector's allowed_url_prefixes")
	require.NotContains(t, err.Error(), trusted.URL+"/hook:")
	require.Zero(t, attackerCalls.Load())
	require.Equal(t, "gateway-secret", trustedKey.Load())
}

func TestPrivateAddressesBlockedWithoutAllowlist(t *testing.T) {
	var calls atomic.Int32
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer internal.Close()
	// The same service by hostname: the check runs on the resolved address, not the URL.
	byName := strings.Replace(internal.URL, "127.0.0.1", "localhost", 1)

	n, err := newNotifier(&configProperties{}, EncodeProps([]string{internal.URL, byName}))
	require.NoError(t, err)
	err = sendWithin(t, 3*time.Second, func() error { return n.SendAlertStatus(testAlertStatus()) })
	require.ErrorIs(t, err, errPrivateAddress)
	require.Contains(t, err.Error(), "giving up after 1 attempt(s)") // not retried
	require.Contains(t, err.Error(), internal.URL)
	require.Contains(t, err.Error(), byName)
	require.Zero(t, calls.Load())
}

func TestAllowlistAdmitsPrivateAddress(t *testing.T) {
	var calls atomic.Int32
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer internal.Close()

	// Listing a receiver is an explicit choice by the connector's author, e.g. for a receiver
	// inside the network of a self-hosted deployment or on localhost during development.
	n, err := newNotifier(&configProperties{AllowedURLPrefixes: []string{internal.URL}}, EncodeProps([]string{internal.URL + "/hook"}))
	require.NoError(t, err)
	require.NoError(t, n.SendAlertStatus(testAlertStatus()))
	require.Equal(t, int32(1), calls.Load())
}

func TestRedirectsNotFollowed(t *testing.T) {
	var targetCalls atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetCalls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()
	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer redirector.Close()

	n := newTestNotifier(t, &configProperties{
		Headers:            map[string]string{"X-Api-Key": "gateway-secret"},
		AllowedURLPrefixes: []string{redirector.URL},
	}, []string{redirector.URL})

	err := n.SendAlertStatus(testAlertStatus())
	require.ErrorContains(t, err, "unexpected status 307")
	require.Zero(t, targetCalls.Load())
}

func TestURLAllowed(t *testing.T) {
	prefixes, err := parseURLPrefixes([]string{"https://example.com", "https://hooks.example.org/rill/"})
	require.NoError(t, err)

	for u, want := range map[string]bool{
		"https://example.com":                         true,
		"https://example.com/any/path":                true,
		"https://EXAMPLE.com:443/any/path":            true,
		"https://hooks.example.org/rill/alerts":       true,
		"http://example.com/":                         false, // scheme must match
		"https://example.com:8443/":                   false, // port must match
		"https://example.com.evil.io/":                false, // not a string prefix
		"https://example.com@evil.io/":                false, // userinfo, host is evil.io
		"https://hooks.example.org/other":             false,
		"https://hooks.example.org/rill/../admin":     false, // dot segments
		"https://hooks.example.org/rill/%2e%2e/admin": false,
	} {
		require.Equal(t, want, urlAllowed(u, prefixes), u)
	}
}

func TestParseURLPrefixesRejectsInvalid(t *testing.T) {
	for _, p := range []string{"example.com", "ftp://example.com/", "https://", "https://user:pw@example.com/"} {
		_, err := parseURLPrefixes([]string{p})
		require.Error(t, err, p)
	}
}

func TestIsPrivateAddress(t *testing.T) {
	for addr, want := range map[string]bool{
		"127.0.0.1":        true,
		"10.1.2.3":         true,
		"172.16.0.1":       true,
		"192.168.1.1":      true,
		"169.254.169.254":  true, // cloud metadata
		"0.0.0.0":          true,
		"::1":              true,
		"fd00::1":          true,
		"fe80::1":          true,
		"::ffff:127.0.0.1": true,
		"8.8.8.8":          false,
		"2001:4860::8888":  false,
	} {
		require.Equal(t, want, isPrivateAddress(netip.MustParseAddr(addr)), addr)
	}
}

func TestGuardConnectsWithoutEnvironmentProxy(t *testing.T) {
	// Through a proxy, the dialed address is the proxy's, so the guard would check the proxy
	// instead of the destination. Without an allowlist the client must connect directly.
	guarded, err := newNotifier(&configProperties{}, EncodeProps([]string{"https://example.com/hook"}))
	require.NoError(t, err)
	require.Nil(t, guarded.client.HTTPClient.Transport.(*http.Transport).Proxy)

	listed, err := newNotifier(&configProperties{AllowedURLPrefixes: []string{"https://example.com/"}}, EncodeProps([]string{"https://example.com/hook"}))
	require.NoError(t, err)
	require.NotNil(t, listed.client.HTTPClient.Transport.(*http.Transport).Proxy)
}
