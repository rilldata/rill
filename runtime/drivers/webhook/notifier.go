package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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
)

type notifier struct {
	signingSecret string
	headers       map[string]string
	props         *NotifierProperties

	// Overridable in tests.
	retryMax       int
	retryWaitMin   time.Duration
	retryWaitMax   time.Duration
	requestTimeout time.Duration
	now            func() time.Time
	newID          func() string
}

type NotifierProperties struct {
	URLs []string `mapstructure:"urls"`
}

func newNotifier(config *configProperties, propsMap map[string]any) (*notifier, error) {
	props, err := DecodeProps(propsMap)
	if err != nil {
		return nil, err
	}
	// Fail fast on a malformed secret instead of erroring on every delivery.
	if _, err := signingKey(config.SigningSecret); err != nil {
		return nil, err
	}
	return &notifier{
		signingSecret:  config.SigningSecret,
		headers:        config.Headers,
		props:          props,
		retryMax:       defaultRetryMax,
		retryWaitMin:   defaultRetryWaitMin,
		retryWaitMax:   defaultRetryWaitMax,
		requestTimeout: defaultRequestTimeout,
		now:            time.Now,
		newID:          uuid.NewString,
	}, nil
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

	client := retryablehttp.NewClient()
	client.Logger = nil
	client.RetryMax = n.retryMax
	client.RetryWaitMin = n.retryWaitMin
	client.RetryWaitMax = n.retryWaitMax
	client.HTTPClient.Timeout = n.requestTimeout

	var errs []error
	for _, u := range dedupe(n.props.URLs) {
		req, err := retryablehttp.NewRequest(http.MethodPost, u, body)
		if err != nil {
			errs = append(errs, fmt.Errorf("webhook %s: %w", u, err))
			continue
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

		resp, err := client.Do(req)
		if err != nil {
			errs = append(errs, fmt.Errorf("webhook %s: %w", u, err))
			continue
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			errs = append(errs, fmt.Errorf("webhook %s: unexpected status %d", u, resp.StatusCode))
		}
	}
	return errors.Join(errs...)
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
