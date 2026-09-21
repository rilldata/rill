package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
)

// payloadVersion identifies the schema of the JSON payload. It must be bumped on any
// backwards-incompatible change to the envelope or event data.
const payloadVersion = 1

const (
	eventTypeAlertStatus     = "alert.status"
	eventTypeScheduledReport = "report.scheduled"
)

const (
	defaultRetryMax       = 2 // 3 attempts in total
	defaultRetryWaitMin   = time.Second
	defaultRetryWaitMax   = 4 * time.Second
	defaultRequestTimeout = 10 * time.Second
	// defaultDeliveryTimeout bounds a whole send, across all URLs and retries. It leaves room
	// for a full retry cycle (3 attempts of up to 10s plus the waits between them).
	defaultDeliveryTimeout = 60 * time.Second
	// maxConcurrentDeliveries bounds how many URLs of one send are delivered to at a time.
	maxConcurrentDeliveries = 8
)

type notifier struct {
	signingSecret   string
	headers         map[string]string
	allowedPrefixes []urlPrefix
	props           *NotifierProperties
	client          *retryablehttp.Client

	// Overridable in tests.
	deliveryTimeout time.Duration
	now             func() time.Time
	newID           func() string
}

type NotifierProperties struct {
	URLs []string `mapstructure:"urls"`
}

func newNotifier(config *configProperties, propsMap map[string]any) (*notifier, error) {
	props, err := DecodeProps(propsMap)
	if err != nil {
		return nil, err
	}
	// Fail fast on a malformed config instead of erroring on every delivery.
	allowedPrefixes, err := validateConfig(config)
	if err != nil {
		return nil, err
	}
	return &notifier{
		signingSecret:   config.SigningSecret,
		headers:         config.Headers,
		allowedPrefixes: allowedPrefixes,
		props:           props,
		// Without an allowlist the URLs are whatever the alert or report author typed,
		// so they must not reach internal services.
		client:          newClient(len(allowedPrefixes) == 0),
		deliveryTimeout: defaultDeliveryTimeout,
		now:             time.Now,
		newID:           uuid.NewString,
	}, nil
}

func newClient(blockPrivate bool) *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.Logger = nil
	client.RetryMax = defaultRetryMax
	client.RetryWaitMin = defaultRetryWaitMin
	client.RetryWaitMax = defaultRetryWaitMax
	client.HTTPClient.Timeout = defaultRequestTimeout
	// The default backoff honors a Retry-After header verbatim. Receivers are arbitrary URLs,
	// so cap the wait at RetryWaitMax like any other backoff.
	client.Backoff = func(minWait, maxWait time.Duration, attempt int, resp *http.Response) time.Duration {
		return min(retryablehttp.DefaultBackoff(minWait, maxWait, attempt, resp), maxWait)
	}
	// The default error on exhausted retries only says how many attempts were made. Keep the
	// final status or error: the execution's error is the only place a failed delivery surfaces.
	client.ErrorHandler = func(resp *http.Response, err error, attempts int) (*http.Response, error) {
		if resp != nil {
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
			_ = resp.Body.Close()
			if err == nil {
				return nil, fmt.Errorf("unexpected status %d after %d attempt(s)", resp.StatusCode, attempts)
			}
		}
		return nil, fmt.Errorf("giving up after %d attempt(s): %w", attempts, err)
	}
	// Webhook receivers don't redirect. Following a redirect would bypass allowed_url_prefixes
	// and carry the static headers to another host, so a 3xx is reported as a failed delivery.
	client.HTTPClient.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	client.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if errors.Is(err, errPrivateAddress) {
			return false, err
		}
		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}
	if blockPrivate {
		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			Control:   blockPrivateAddresses,
		}
		transport := client.HTTPClient.Transport.(*http.Transport)
		transport.DialContext = dialer.DialContext
		// Connect directly: through a proxy, the guard would check the proxy's address, and the
		// proxy would reach the destination unchecked.
		transport.Proxy = nil
	}
	return client
}

// validateConfig checks the connector config and returns the parsed allowed_url_prefixes.
func validateConfig(config *configProperties) ([]urlPrefix, error) {
	if _, err := signingKey(config.SigningSecret); err != nil {
		return nil, err
	}
	prefixes, err := parseURLPrefixes(config.AllowedURLPrefixes)
	if err != nil {
		return nil, err
	}
	// Anyone who can create an alert or report chooses its URLs, so static headers (typically
	// credentials) are only allowed when the connector restricts where deliveries can go.
	if len(config.Headers) > 0 && len(prefixes) == 0 {
		return nil, errors.New("webhook connector: headers require allowed_url_prefixes, so they are only sent to the listed URLs")
	}
	return prefixes, nil
}

func EncodeProps(urls []string) map[string]any {
	return map[string]any{
		"urls": pbutil.ToSliceAny(urls),
	}
}

func DecodeProps(propsMap map[string]any) (*NotifierProperties, error) {
	props := &NotifierProperties{}
	err := mapstructure.WeakDecode(propsMap, props)
	if err != nil {
		return nil, err
	}
	return props, nil
}

// payload is the versioned envelope sent to every URL.
type payload struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Version   int       `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
}

type alertStatusData struct {
	DisplayName    string         `json:"display_name"`
	ExecutionTime  time.Time      `json:"execution_time"`
	Status         string         `json:"status"` // PASS, FAIL or ERROR
	IsRecover      bool           `json:"is_recover"`
	FailRow        map[string]any `json:"fail_row,omitempty"`
	ExecutionError string         `json:"execution_error,omitempty"`
	OpenLink       string         `json:"open_link,omitempty"`
	EditLink       string         `json:"edit_link,omitempty"`
}

type scheduledReportData struct {
	DisplayName    string    `json:"display_name"`
	ReportTime     time.Time `json:"report_time"`
	DownloadFormat string    `json:"download_format,omitempty"`
	Summary        string    `json:"summary,omitempty"`
	OpenLink       string    `json:"open_link,omitempty"`
	DownloadLink   string    `json:"download_link,omitempty"`
}

func (n *notifier) SendAlertStatus(s *drivers.AlertStatus) error {
	var status string
	switch s.Status {
	case runtimev1.AssertionStatus_ASSERTION_STATUS_PASS:
		status = "PASS"
	case runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL:
		status = "FAIL"
	case runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR:
		status = "ERROR"
	default:
		return fmt.Errorf("unknown assertion status: %v", s.Status)
	}

	return n.send(eventTypeAlertStatus, &alertStatusData{
		DisplayName:    s.DisplayName,
		ExecutionTime:  s.ExecutionTime,
		Status:         status,
		IsRecover:      s.IsRecover,
		FailRow:        s.FailRow,
		ExecutionError: s.ExecutionError,
		OpenLink:       s.OpenLink,
		EditLink:       s.EditLink,
	})
}

func (n *notifier) SendScheduledReport(s *drivers.ScheduledReport) error {
	return n.send(eventTypeScheduledReport, &scheduledReportData{
		DisplayName:    s.DisplayName,
		ReportTime:     s.ReportTime,
		DownloadFormat: s.DownloadFormat,
		Summary:        s.Summary,
		OpenLink:       s.OpenLink,
		DownloadLink:   s.DownloadLink,
	})
}

// send delivers the event to all configured URLs. Every URL is attempted even if earlier
// ones fail. Because Rill keeps no per-delivery log (only the execution's error message),
// each per-URL error must be self-sufficient: it names the URL and the final outcome.
func (n *notifier) send(eventType string, data any) error {
	id := n.newID()
	ts := n.now().UTC()

	body, err := json.Marshal(&payload{
		ID:        id,
		Type:      eventType,
		Version:   payloadVersion,
		Timestamp: ts,
		Data:      data,
	})
	if err != nil {
		return fmt.Errorf("webhook: failed to encode payload: %w", err)
	}

	var signature string
	if n.signingSecret != "" {
		signature, err = sign(n.signingSecret, id, ts, body)
		if err != nil {
			return fmt.Errorf("webhook: %w", err)
		}
	}

	// Deliver to the URLs in parallel (up to maxConcurrentDeliveries at a time) under one deadline,
	// so a slow receiver neither delays the others nor holds the caller for longer than deliveryTimeout.
	ctx, cancel := context.WithTimeout(context.Background(), n.deliveryTimeout)
	defer cancel()

	urls := dedupe(n.props.URLs)
	errs := make([]error, len(urls))
	sem := make(chan struct{}, maxConcurrentDeliveries)
	var wg sync.WaitGroup
	for i, u := range urls {
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			errs[i] = fmt.Errorf("webhook %s: %w", u, ctx.Err())
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			errs[i] = n.deliver(ctx, u, id, ts, body, signature)
		}()
	}
	wg.Wait()

	// The reconcilers build a notifier per execution, so keeping connections alive would only
	// leave them open until the transport's idle timeout.
	n.client.HTTPClient.CloseIdleConnections()

	return errors.Join(errs...)
}

// deliver posts the payload to a single URL. The returned error names the URL and the final outcome.
func (n *notifier) deliver(ctx context.Context, u, id string, ts time.Time, body []byte, signature string) error {
	if len(n.allowedPrefixes) > 0 && !urlAllowed(u, n.allowedPrefixes) {
		return fmt.Errorf("webhook %s: URL is not in the connector's allowed_url_prefixes", u)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, u, body)
	if err != nil {
		return fmt.Errorf("webhook %s: %w", u, err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range n.headers {
		req.Header.Set(k, v)
	}
	if signature != "" {
		req.Header.Set("webhook-id", id)
		req.Header.Set("webhook-timestamp", strconv.FormatInt(ts.Unix(), 10))
		req.Header.Set("webhook-signature", signature)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook %s: %w", u, err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook %s: unexpected status %d", u, resp.StatusCode)
	}
	return nil
}

// sign produces a signature following the Standard Webhooks specification:
// base64 HMAC-SHA256 over "{id}.{timestamp}.{body}" with a "v1," prefix.
func sign(secret, id string, ts time.Time, body []byte) (string, error) {
	key, err := signingKey(secret)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	fmt.Fprintf(mac, "%s.%d.", id, ts.Unix())
	mac.Write(body)
	return "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

func signingKey(secret string) ([]byte, error) {
	rest, ok := strings.CutPrefix(secret, "whsec_")
	if !ok {
		return []byte(secret), nil
	}
	key, err := base64.StdEncoding.DecodeString(rest)
	if err != nil {
		// Deliberately omits the secret value: this error surfaces in logs and execution history.
		return nil, errors.New("invalid signing secret: not valid base64 after the whsec_ prefix")
	}
	return key, nil
}

func dedupe(urls []string) []string {
	seen := make(map[string]bool, len(urls))
	res := make([]string, 0, len(urls))
	for _, u := range urls {
		if !seen[u] {
			seen[u] = true
			res = append(res, u)
		}
	}
	return res
}

// urlPrefix is a parsed entry of allowed_url_prefixes.
type urlPrefix struct {
	scheme string
	host   string // host:port, with the scheme's default port filled in
	path   string
}

func parseURLPrefixes(prefixes []string) ([]urlPrefix, error) {
	res := make([]urlPrefix, 0, len(prefixes))
	for _, p := range prefixes {
		u, err := url.Parse(p)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
			return nil, fmt.Errorf("webhook connector: invalid allowed_url_prefixes entry %q: must be an absolute http or https URL", p)
		}
		res = append(res, urlPrefix{scheme: u.Scheme, host: canonicalHost(u), path: u.Path})
	}
	return res, nil
}

// urlAllowed reports whether rawURL falls under one of the prefixes. Unlike a plain string
// prefix, the scheme and host must match exactly, so https://example.com does not admit
// https://example.com.evil.io, and paths with dot segments are rejected because the receiver
// may resolve them outside the prefix.
func urlAllowed(rawURL string, prefixes []urlPrefix) bool {
	u, err := url.Parse(rawURL)
	if err != nil || u.User != nil || slices.ContainsFunc(strings.Split(u.Path, "/"), func(seg string) bool {
		return seg == "." || seg == ".."
	}) {
		return false
	}
	host := canonicalHost(u)
	for _, p := range prefixes {
		if u.Scheme == p.scheme && host == p.host && strings.HasPrefix(u.Path, p.path) {
			return true
		}
	}
	return false
}

func canonicalHost(u *url.URL) string {
	port := u.Port()
	if port == "" {
		port = "80"
		if u.Scheme == "https" {
			port = "443"
		}
	}
	return net.JoinHostPort(strings.ToLower(u.Hostname()), port)
}

var errPrivateAddress = errors.New("destination is not a public address")

// blockPrivateAddresses is a net.Dialer Control function. It runs after DNS resolution, on the
// address actually being dialed, so a hostname that resolves to an internal address is caught too.
func blockPrivateAddresses(_, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return err
	}
	if isPrivateAddress(ip) {
		return fmt.Errorf("%w: %s", errPrivateAddress, ip)
	}
	return nil
}

// isPrivateAddress reports whether ip is loopback, private (RFC 1918 or IPv6 unique local),
// link-local (which includes cloud metadata endpoints such as 169.254.169.254) or unspecified.
func isPrivateAddress(ip netip.Addr) bool {
	ip = ip.Unmap()
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}
