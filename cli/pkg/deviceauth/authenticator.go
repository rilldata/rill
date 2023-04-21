package deviceauth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/pkg/errors"
	"github.com/rilldata/rill/admin/database"
)

// Most parts of this file are copied from https://github.com/planetscale/cli/blob/main/internal/auth/authenticator.go

const (
	formMediaType = "application/x-www-form-urlencoded"
	jsonMediaType = "application/json"
)

// Authenticator is the interface for authentication via device oauth
type Authenticator interface {
	VerifyDevice(ctx context.Context) (*DeviceVerification, error)
	GetAccessTokenForDevice(ctx context.Context, v DeviceVerification) (string, error)
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
	client   *http.Client
	BaseURL  *url.URL
	Clock    clock.Clock
	ClientID string
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
		"redirect":  []string{url.QueryEscape(redirectURL)},
	})
	if err != nil {
		return nil, err
	}

	res, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if _, err = checkErrorResponse(res); err != nil {
		return nil, err
	}

	deviceCodeRes := &DeviceCodeResponse{}
	err = json.NewDecoder(res.Body).Decode(deviceCodeRes)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding device code response")
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
func (d *DeviceAuthenticator) GetAccessTokenForDevice(ctx context.Context, v *DeviceVerification) (*OAuthTokenResponse, error) {
	for {
		// This loop begins right after we open the user's browser to send an
		// authentication code. We don't request a token immediately because the
		// has to complete that authentication flow before we can provide a
		// token anyway.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(v.CheckInterval):
			// Ready to check again.
		}

		token, err := d.requestToken(ctx, v.DeviceCode, d.ClientID)
		if err != nil {
			// Fatal error.
			return nil, err
		}
		if token != nil {
			// Successful authentication.
			return token, nil
		}

		if d.Clock.Now().After(v.ExpiresAt) {
			return nil, errors.New("authentication timed out")
		}
	}
}

// OAuthTokenResponse contains the information returned after fetching an access token for a device.
type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in,string"`
	TokenType   string `json:"token_type"`
	UserID      string `json:"user_id"`
}

func (d *DeviceAuthenticator) requestToken(ctx context.Context, deviceCode, clientID string) (*OAuthTokenResponse, error) {
	req, err := d.newFormRequest(ctx, "auth/oauth/token", url.Values{
		"grant_type":  []string{"urn:ietf:params:oauth:grant-type:device_code"},
		"device_code": []string{deviceCode},
		"client_id":   []string{clientID},
	})
	if err != nil {
		return nil, errors.Wrap(err, "error creating request")
	}

	res, err := d.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error performing http request")
	}
	defer res.Body.Close()

	isRetryable, err := checkErrorResponse(res)
	if err != nil {
		return nil, err
	}

	// Bail early so the token fetching is retried.
	if isRetryable {
		return nil, nil
	}

	tokenRes := &OAuthTokenResponse{}
	err = json.NewDecoder(res.Body).Decode(tokenRes)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding token response")
	}

	return tokenRes, nil
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

	req.Header.Set("Content-Type", formMediaType)
	req.Header.Set("Accept", jsonMediaType)
	return req, nil
}

// checkErrorResponse returns whether the error is retryable or not and the error itself.
func checkErrorResponse(res *http.Response) (bool, error) {
	if res.StatusCode < http.StatusBadRequest {
		// 200 OK, etc.
		return false, nil
	}

	// Client or server error.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	bodyStr := string(bytes.TrimSpace(body))
	// If we're polling and haven't authorized yet or we need to slow down, we don't want to terminate the polling
	if bodyStr == "authorization_pending" || bodyStr == "slow_down" {
		return true, nil
	}
	if bodyStr == "expired_token" || bodyStr == "rejected" {
		return false, errors.New(bodyStr)
	}

	return false, errors.New(strconv.Itoa(res.StatusCode) + ": " + bodyStr)
}
