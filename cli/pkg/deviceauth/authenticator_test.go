package deviceauth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/rilldata/rill/admin/pkg/oauth"
	"github.com/stretchr/testify/require"
)

type requestSnapshot struct {
	method      string
	path        string
	contentType string
	accept      string
	form        url.Values
	parseErr    error
}

var _ Authenticator = (*DeviceAuthenticator)(nil)

func TestVerifyDeviceContract(t *testing.T) {
	t.Parallel()
	// Verify the initial device-authorization form and its conversion into deterministic polling state.

	requests := make(chan requestSnapshot, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- snapshotRequest(r)
		w.Header().Set("Content-Type", oauth.JSONMediaType)
		_, _ = io.WriteString(w, `{
			"device_code":"device-code",
			"user_code":"ABCD-EFGH",
			"verification_uri":"https://auth.example/verify",
			"verification_uri_complete":"https://auth.example/verify?user_code=ABCD-EFGH",
			"expires_in":60,
			"interval":2
		}`)
	}))
	t.Cleanup(server.Close)

	mockClock := clock.NewMock()
	authenticator := newTestAuthenticator(t, server, mockClock)
	redirectURL := "http://localhost:9009/auth/callback?project=my project&open=true"
	verification, err := authenticator.VerifyDevice(context.Background(), redirectURL)
	require.NoError(t, err)

	request := <-requests
	require.NoError(t, request.parseErr)
	require.Equal(t, http.MethodPost, request.method)
	require.Equal(t, "/auth/oauth/device_authorization", request.path)
	require.Equal(t, oauth.FormMediaType, request.contentType)
	require.Equal(t, oauth.JSONMediaType, request.accept)
	require.Equal(t, url.Values{
		"client_id": {authenticator.ClientID},
		"scope":     {"full_account"},
		"redirect":  {redirectURL},
	}, request.form)
	require.Equal(t, &DeviceVerification{
		DeviceCode:              "device-code",
		UserCode:                "ABCD-EFGH",
		VerificationURL:         "https://auth.example/verify",
		VerificationCompleteURL: "https://auth.example/verify?user_code=ABCD-EFGH",
		CheckInterval:           2 * time.Second,
		ExpiresAt:               mockClock.Now().Add(time.Minute),
	}, verification)
}

func TestVerifyDeviceFailures(t *testing.T) {
	t.Parallel()
	// Malformed and unsuccessful authorization responses must remain errors rather than partial device sessions.

	tests := []struct {
		name        string
		status      int
		body        string
		wantErrText string
	}{
		{
			name:        "malformed JSON",
			status:      http.StatusOK,
			body:        `{"device_code":`,
			wantErrText: "error decoding device code response",
		},
		{
			name:        "server error",
			status:      http.StatusInternalServerError,
			body:        "backend unavailable",
			wantErrText: "500: backend unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.body)
			}))
			t.Cleanup(server.Close)

			authenticator := newTestAuthenticator(t, server, clock.NewMock())
			verification, err := authenticator.VerifyDevice(context.Background(), "http://localhost/callback")
			require.ErrorContains(t, err, tt.wantErrText)
			require.Nil(t, verification)
		})
	}
}

func TestGetAccessTokenForDeviceContract(t *testing.T) {
	t.Parallel()
	// Model the complete polling protocol, including backoff, expiry, terminal errors, and credential persistence.

	type reply struct {
		status int
		body   string
	}
	tests := []struct {
		name          string
		replies       []reply
		interval      time.Duration
		expiresIn     time.Duration
		wantWaits     []time.Duration
		wantRequests  int
		wantToken     string
		wantErrIs     error
		wantErrText   string
		wantPersisted string
	}{
		{
			name: "authorization pending then success",
			replies: []reply{
				{status: http.StatusUnauthorized, body: "authorization_pending"},
				{status: http.StatusOK, body: `{"access_token":"new-token","token_type":"Bearer"}`},
			},
			interval:      2 * time.Second,
			expiresIn:     time.Minute,
			wantWaits:     []time.Duration{2 * time.Second, 2 * time.Second},
			wantRequests:  2,
			wantToken:     "new-token",
			wantPersisted: "new-token",
		},
		{
			name: "slow down increases all later intervals",
			replies: []reply{
				{status: http.StatusBadRequest, body: "slow_down"},
				{status: http.StatusUnauthorized, body: "authorization_pending"},
				{status: http.StatusOK, body: `{"access_token":"new-token","token_type":"Bearer"}`},
			},
			interval:      2 * time.Second,
			expiresIn:     time.Minute,
			wantWaits:     []time.Duration{2 * time.Second, 7 * time.Second, 7 * time.Second},
			wantRequests:  3,
			wantToken:     "new-token",
			wantPersisted: "new-token",
		},
		{
			name: "rejection is fatal",
			replies: []reply{
				{status: http.StatusUnauthorized, body: "rejected"},
				{status: http.StatusOK, body: `{"access_token":"must-not-be-read"}`},
			},
			interval:      2 * time.Second,
			expiresIn:     time.Minute,
			wantWaits:     []time.Duration{2 * time.Second},
			wantRequests:  1,
			wantErrIs:     ErrCodeRejected,
			wantPersisted: "old-token",
		},
		{
			name: "server reports expiry",
			replies: []reply{
				{status: http.StatusUnauthorized, body: "expired_token"},
				{status: http.StatusOK, body: `{"access_token":"must-not-be-read"}`},
			},
			interval:      2 * time.Second,
			expiresIn:     time.Minute,
			wantWaits:     []time.Duration{2 * time.Second},
			wantRequests:  1,
			wantErrIs:     ErrAuthenticationTimedout,
			wantPersisted: "old-token",
		},
		{
			name: "local expiry caps the final wait",
			replies: []reply{
				{status: http.StatusUnauthorized, body: "authorization_pending"},
				{status: http.StatusOK, body: `{"access_token":"must-not-be-read"}`},
			},
			interval:      2 * time.Second,
			expiresIn:     3 * time.Second,
			wantWaits:     []time.Duration{2 * time.Second, time.Second},
			wantRequests:  1,
			wantErrIs:     ErrAuthenticationTimedout,
			wantPersisted: "old-token",
		},
		{
			name: "malformed token JSON is fatal",
			replies: []reply{
				{status: http.StatusOK, body: `{"access_token":`},
				{status: http.StatusOK, body: `{"access_token":"must-not-be-read"}`},
			},
			interval:      2 * time.Second,
			expiresIn:     time.Minute,
			wantWaits:     []time.Duration{2 * time.Second},
			wantRequests:  1,
			wantErrText:   "error decoding token response",
			wantPersisted: "old-token",
		},
		{
			name: "empty token response is not success",
			replies: []reply{
				{status: http.StatusOK, body: `{}`},
				{status: http.StatusOK, body: `{"access_token":"must-not-be-read"}`},
			},
			interval:      2 * time.Second,
			expiresIn:     time.Minute,
			wantWaits:     []time.Duration{2 * time.Second},
			wantRequests:  1,
			wantErrText:   "missing access_token",
			wantPersisted: "old-token",
		},
		{
			name: "server error is fatal",
			replies: []reply{
				{status: http.StatusInternalServerError, body: "backend unavailable"},
				{status: http.StatusOK, body: `{"access_token":"must-not-be-read"}`},
			},
			interval:      2 * time.Second,
			expiresIn:     time.Minute,
			wantWaits:     []time.Duration{2 * time.Second},
			wantRequests:  1,
			wantErrText:   "500: backend unavailable",
			wantPersisted: "old-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var persisted atomic.Value
			persisted.Store("old-token")
			var changedBeforeSuccess atomic.Bool
			requests := make(chan requestSnapshot, len(tt.replies)+1)
			var requestCount atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if persisted.Load().(string) != "old-token" {
					changedBeforeSuccess.Store(true)
				}
				requests <- snapshotRequest(r)
				index := int(requestCount.Add(1)) - 1
				if index >= len(tt.replies) {
					http.Error(w, "unexpected extra poll", http.StatusInternalServerError)
					return
				}
				reply := tt.replies[index]
				w.WriteHeader(reply.status)
				_, _ = io.WriteString(w, reply.body)
			}))
			t.Cleanup(server.Close)

			mockClock := clock.NewMock()
			authenticator := newTestAuthenticator(t, server, mockClock)
			var waits []time.Duration
			authenticator.waitForNextPoll = func(ctx context.Context, interval time.Duration) error {
				if err := ctx.Err(); err != nil {
					return err
				}
				waits = append(waits, interval)
				mockClock.Add(interval)
				return nil
			}
			verification := &DeviceVerification{
				DeviceCode:    "device-code",
				CheckInterval: tt.interval,
				ExpiresAt:     mockClock.Now().Add(tt.expiresIn),
			}

			token, err := authenticator.GetAccessTokenForDevice(context.Background(), verification)
			// This mirrors the CLI persistence boundary: no local credential is
			// replaced until polling returns a complete, usable token.
			if err == nil {
				persisted.Store(token.AccessToken)
			}

			if tt.wantErrIs != nil {
				require.ErrorIs(t, err, tt.wantErrIs)
			} else if tt.wantErrText != "" {
				require.ErrorContains(t, err, tt.wantErrText)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantToken, token.AccessToken)
			}
			if err != nil {
				require.Nil(t, token)
			}
			require.Equal(t, tt.wantWaits, waits)
			require.Equal(t, int32(tt.wantRequests), requestCount.Load())
			require.Equal(t, tt.wantPersisted, persisted.Load())
			require.False(t, changedBeforeSuccess.Load())

			close(requests)
			for request := range requests {
				require.NoError(t, request.parseErr)
				require.Equal(t, http.MethodPost, request.method)
				require.Equal(t, "/auth/oauth/token", request.path)
				require.Equal(t, oauth.FormMediaType, request.contentType)
				require.Equal(t, oauth.JSONMediaType, request.accept)
				require.Equal(t, url.Values{
					"grant_type":             {"urn:ietf:params:oauth:grant-type:device_code"},
					"device_code":            {"device-code"},
					"client_id":              {authenticator.ClientID},
					"token_response_version": {"standard"},
				}, request.form)
			}
		})
	}
}

func TestGetAccessTokenForDeviceCancellation(t *testing.T) {
	t.Parallel()
	// Cancellation must stop both a pending wait and an already in-flight token request without leaking a token.

	t.Run("between polls", func(t *testing.T) {
		t.Parallel()

		var requestCount atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			requestCount.Add(1)
		}))
		t.Cleanup(server.Close)

		mockClock := clock.NewMock()
		authenticator := newTestAuthenticator(t, server, mockClock)
		ctx, cancel := context.WithCancel(context.Background())
		authenticator.waitForNextPoll = func(ctx context.Context, _ time.Duration) error {
			cancel()
			<-ctx.Done()
			return ctx.Err()
		}

		token, err := authenticator.GetAccessTokenForDevice(ctx, &DeviceVerification{
			DeviceCode:    "device-code",
			CheckInterval: time.Minute,
			ExpiresAt:     mockClock.Now().Add(time.Hour),
		})
		require.ErrorIs(t, err, context.Canceled)
		require.Nil(t, token)
		require.Zero(t, requestCount.Load())
	})

	t.Run("in flight", func(t *testing.T) {
		t.Parallel()

		requestStarted := make(chan struct{})
		releaseRequest := make(chan struct{})
		server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			close(requestStarted)
			<-releaseRequest
		}))
		t.Cleanup(server.Close)
		t.Cleanup(func() { close(releaseRequest) })

		mockClock := clock.NewMock()
		authenticator := newTestAuthenticator(t, server, mockClock)
		authenticator.waitForNextPoll = func(_ context.Context, interval time.Duration) error {
			mockClock.Add(interval)
			return nil
		}
		ctx, cancel := context.WithCancel(context.Background())
		type result struct {
			token *oauth.TokenResponse
			err   error
		}
		resultCh := make(chan result, 1)
		go func() {
			token, err := authenticator.GetAccessTokenForDevice(ctx, &DeviceVerification{
				DeviceCode:    "device-code",
				CheckInterval: time.Second,
				ExpiresAt:     mockClock.Now().Add(time.Hour),
			})
			resultCh <- result{token: token, err: err}
		}()

		<-requestStarted
		cancel()
		select {
		case result := <-resultCh:
			require.ErrorIs(t, result.err, context.Canceled)
			require.Nil(t, result.token)
		case <-time.After(2 * time.Second):
			t.Fatal("token request did not stop after cancellation")
		}
	})
}

func newTestAuthenticator(t *testing.T, server *httptest.Server, mockClock *clock.Mock) *DeviceAuthenticator {
	t.Helper()

	authenticator, err := New(server.URL + "/")
	require.NoError(t, err)
	authenticator.client = server.Client()
	authenticator.Clock = mockClock
	return authenticator
}

func snapshotRequest(r *http.Request) requestSnapshot {
	err := r.ParseForm()
	return requestSnapshot{
		method:      r.Method,
		path:        r.URL.Path,
		contentType: r.Header.Get("Content-Type"),
		accept:      r.Header.Get("Accept"),
		form:        r.PostForm,
		parseErr:    err,
	}
}
