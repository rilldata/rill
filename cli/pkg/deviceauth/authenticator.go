package deviceauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/oauth"
)

// Most parts of this file are copied from https://github.com/planetscale/cli/blob/main/internal/auth/authenticator.go

var (
	ErrAuthenticationTimedout = fmt.Errorf("authentication timed out")
	ErrCodeRejected           = fmt.Errorf("confirmation code rejected")
)

const slowDownBackoff = 5 * time.Second

type pollAction int

const (
	pollComplete pollAction = iota
	pollPending
	pollSlowDown
)

// Authenticator is the interface for authentication via device oauth
type Authenticator interface {
	VerifyDevice(ctx context.Context, redirectURL string) (*DeviceVerification, error)
	GetAccessTokenForDevice(ctx context.Context, v *DeviceVerification) (*oauth.TokenResponse, error)
}

// DeviceCodeResponse encapsulates the response for obtaining a device code.
type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationCompleteURI string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	PollingInterval         int    `json:"interval"`
}

// DeviceVerification represents the response from verifying a device.
type DeviceVerification struct {
	DeviceCode              string
	UserCode                string
	VerificationURL         string
	VerificationCompleteURL string
	CheckInterval           time.Duration
	ExpiresAt               time.Time
}

// DeviceAuthenticator performs the authentication flow for logging in.
type DeviceAuthenticator struct {
	client          *http.Client
	waitForNextPoll func(context.Context, time.Duration) error
	BaseURL         *url.URL
	Clock           clock.Clock
	ClientID        string
}

// New returns an instance of the DeviceAuthenticator
func New(authURL string) (*DeviceAuthenticator, error) {
	baseURL, err := url.Parse(authURL)
	if err != nil {
		return nil, err
	}

	authenticator := &DeviceAuthenticator{
		client:   http.DefaultClient,
		BaseURL:  baseURL,
		Clock:    clock.New(),
		ClientID: database.AuthClientIDRillCLI,
	}

	return authenticator, nil
}

// VerifyDevice performs the device verification API calls.
func (d *DeviceAuthenticator) VerifyDevice(ctx context.Context, redirectURL string) (*DeviceVerification, error) {
	req, err := d.newFormRequest(ctx, "auth/oauth/device_authorization", url.Values{
		"client_id": []string{d.ClientID},
		"scope":     []string{"full_account"},
		"redirect":  []string{redirectURL},
	})
	if err != nil {
		return nil, err
	}

	res, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	action, err := checkErrorResponse(res)
	if err != nil {
		return nil, err
	}
	if action != pollComplete {
		return nil, errors.New("unexpected retry response while requesting a device code")
	}

	deviceCodeRes := &DeviceCodeResponse{}
	err = json.NewDecoder(res.Body).Decode(deviceCodeRes)
	if err != nil {
		return nil, fmt.Errorf("error decoding device code response: %w", err)
	}

	checkInterval := time.Duration(deviceCodeRes.PollingInterval) * time.Second
	if checkInterval == 0 {
		checkInterval = time.Duration(5) * time.Second
	}

	expiresAt := d.Clock.Now().Add(time.Duration(deviceCodeRes.ExpiresIn) * time.Second)

	return &DeviceVerification{
		DeviceCode:              deviceCodeRes.DeviceCode,
		UserCode:                deviceCodeRes.UserCode,
		VerificationCompleteURL: deviceCodeRes.VerificationCompleteURI,
		VerificationURL:         deviceCodeRes.VerificationURI,
		ExpiresAt:               expiresAt,
		CheckInterval:           checkInterval,
	}, nil
}

// GetAccessTokenForDevice uses the device verification response to fetch an access token.
func (d *DeviceAuthenticator) GetAccessTokenForDevice(ctx context.Context, v *DeviceVerification) (*oauth.TokenResponse, error) {
	pollInterval := v.CheckInterval
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// Poll only after the server-provided interval. The wait is capped at the
		// device-code expiry so a long interval cannot keep login alive past it.
		now := d.Clock.Now()
		if !now.Before(v.ExpiresAt) {
			return nil, ErrAuthenticationTimedout
		}
		waitFor := pollInterval
		if remaining := v.ExpiresAt.Sub(now); remaining < waitFor {
			waitFor = remaining
		}
		wait := d.waitForNextPoll
		if wait == nil {
			wait = d.wait
		}
		if err := wait(ctx, waitFor); err != nil {
			return nil, err
		}
		if !d.Clock.Now().Before(v.ExpiresAt) {
			return nil, ErrAuthenticationTimedout
		}

		token, action, err := d.requestToken(ctx, v.DeviceCode, d.ClientID)
		if err != nil {
			return nil, err
		}
		switch action {
		case pollComplete:
			// A token is exposed only after the complete success response has been
			// decoded and validated. Callers can persist it after this return without
			// risking a credential change for a pending or partial response.
			return token, nil
		case pollSlowDown:
			// RFC 8628 requires every poll after slow_down to use an interval that
			// is at least five seconds longer than the previous one.
			pollInterval += slowDownBackoff
		case pollPending:
			// The user has not completed authorization yet; keep the current interval.
		}
	}
}

// wait keeps cancellation responsive while the poller is between HTTP requests.
// It is a field-backed method so tests can substitute a deterministic sleeper.
func (d *DeviceAuthenticator) wait(ctx context.Context, interval time.Duration) error {
	timer := d.Clock.Timer(interval)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (d *DeviceAuthenticator) requestToken(ctx context.Context, deviceCode, clientID string) (*oauth.TokenResponse, pollAction, error) {
	req, err := d.newFormRequest(ctx, "auth/oauth/token", url.Values{
		"grant_type":             []string{"urn:ietf:params:oauth:grant-type:device_code"},
		"device_code":            []string{deviceCode},
		"client_id":              []string{clientID},
		"token_response_version": []string{"standard"}, // For backward compatibility with older Rill CLI, see utils.go in oauth pkg
	})
	if err != nil {
		return nil, pollComplete, fmt.Errorf("error creating request: %w", err)
	}

	res, err := d.client.Do(req)
	if err != nil {
		return nil, pollComplete, fmt.Errorf("error performing http request: %w", err)
	}
	defer res.Body.Close()

	action, err := checkErrorResponse(res)
	if err != nil {
		return nil, pollComplete, err
	}

	if action != pollComplete {
		return nil, action, nil
	}

	tokenRes := &oauth.TokenResponse{}
	err = json.NewDecoder(res.Body).Decode(tokenRes)
	if err != nil {
		return nil, pollComplete, fmt.Errorf("error decoding token response: %w", err)
	}
	if tokenRes.AccessToken == "" {
		return nil, pollComplete, errors.New("token response is missing access_token")
	}

	return tokenRes, pollComplete, nil
}

// newFormRequest creates a new form URL encoded request
func (d *DeviceAuthenticator) newFormRequest(ctx context.Context, path string, payload url.Values) (*http.Request, error) {
	u, err := d.BaseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	// Emulate the format of data sent by http.Client's PostForm method, but
	// also preserve context support.
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		u.String(),
		strings.NewReader(payload.Encode()),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", oauth.FormMediaType)
	req.Header.Set("Accept", oauth.JSONMediaType)
	return req, nil
}

// checkErrorResponse classifies the two retry instructions used by device polling.
func checkErrorResponse(res *http.Response) (pollAction, error) {
	if res.StatusCode < http.StatusBadRequest {
		return pollComplete, nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return pollComplete, err
	}
	bodyStr := string(bytes.TrimSpace(body))
	if res.StatusCode < http.StatusInternalServerError {
		switch bodyStr {
		case "authorization_pending":
			return pollPending, nil
		case "slow_down":
			return pollSlowDown, nil
		case "expired_token":
			return pollComplete, ErrAuthenticationTimedout
		case "rejected":
			return pollComplete, ErrCodeRejected
		}
	}

	return pollComplete, errors.New(strconv.Itoa(res.StatusCode) + ": " + bodyStr)
}
