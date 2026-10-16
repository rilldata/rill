package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/oauthclient"
	"golang.org/x/oauth2"
)

// This file implements the OAuth state the driver keeps for Rill Developer, where there is a single local user.
// The OAuth client itself (discovery, registration, PKCE, refresh) is implemented in runtime/pkg/oauthclient.

// ErrOAuthNotConfigured is returned when an OAuth operation is attempted on a server without auth "oauth".
var ErrOAuthNotConfigured = errors.New("the connector does not use OAuth (set 'auth: oauth' to enable it)")

// oauthStateFilename is the file in the connector's data dir that holds locally persisted OAuth state (Rill Developer).
const oauthStateFilename = "oauth.json"

// oauthConfigHash identifies a connector's OAuth configuration: the server URL, the pre-registered client and the scopes.
// Authorizations are bound to it, in Rill Developer and in Rill Cloud, so changing any of them requires the user to authorize again.
// Each field is length-prefixed so that different configurations can never produce the same input.
func oauthConfigHash(cfg *configProperties) string {
	h := sha256.New()
	for _, s := range append([]string{cfg.URL, cfg.OAuthClientID, cfg.OAuthClientSecret}, cfg.OAuthScopes...) {
		_, _ = fmt.Fprintf(h, "%d:%s;", len(s), s)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// OAuthState is the OAuth state Rill Developer persists for a connector: the client and the local user's token.
type OAuthState struct {
	// ConfigHash is the oauthConfigHash of the configuration the state was established for; state for another configuration is discarded.
	ConfigHash string              `json:"config_hash"`
	Client     *oauthclient.Client `json:"client"`
	Token      *oauth2.Token       `json:"token,omitempty"`
	// TokenObtainedOn is when the token was issued or last refreshed.
	TokenObtainedOn time.Time `json:"token_obtained_on"`
}

// localOAuth manages the OAuth state persisted in the connector's data dir.
// Rill Developer has a single user, so the token is stored alongside the client; in Rill Cloud the admin service holds tokens per user instead.
type localOAuth struct {
	h      *handle
	mu     sync.Mutex
	loaded bool
	state  *OAuthState
}

// load returns the persisted state, or nil if there is none.
func (l *localOAuth) load() (*OAuthState, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.loadLocked()
}

func (l *localOAuth) loadLocked() (*OAuthState, error) {
	if l.loaded {
		return l.state, nil
	}
	path, err := l.path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			l.loaded = true
			return nil, nil
		}
		return nil, err
	}
	state := &OAuthState{}
	if err := json.Unmarshal(data, state); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	// Discard state established for a different OAuth configuration.
	if state.Client == nil || state.ConfigHash != l.h.oauthConfigHash {
		state = nil
	}
	l.state = state
	l.loaded = true
	return state, nil
}

func (l *localOAuth) save(state *OAuthState) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.saveLocked(state)
}

func (l *localOAuth) saveLocked(state *OAuthState) error {
	path, err := l.path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return err
	}
	l.state = state
	l.loaded = true
	return nil
}

func (l *localOAuth) clear() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	path, err := l.path()
	if err != nil {
		return err
	}
	l.state = nil
	l.loaded = true
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (l *localOAuth) path() (string, error) {
	dir, err := l.h.storage.DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, oauthStateFilename), nil
}

// accessToken returns a valid access token from the persisted state, refreshing it when it is about to expire.
func (l *localOAuth) accessToken(ctx context.Context, forceRefresh bool) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	state, err := l.loadLocked()
	if err != nil {
		return "", err
	}
	if state == nil || state.Token == nil || state.Token.AccessToken == "" {
		return "", drivers.ErrMCPAuthorizationRequired
	}
	tok := state.Token
	expiring := oauthclient.TokenExpiring(tok, state.TokenObtainedOn)
	if !forceRefresh && !expiring {
		return tok.AccessToken, nil
	}
	if tok.RefreshToken == "" {
		// The token can't be renewed early, so it is used until it actually expires.
		if tok.Valid() {
			return tok.AccessToken, nil
		}
		return "", fmt.Errorf("%w: the token has expired", drivers.ErrMCPAuthorizationRequired)
	}
	newTok, err := state.Client.Refresh(ctx, tok, l.h.httpClient())
	if err != nil {
		return "", err
	}
	state.Token = newTok
	state.TokenObtainedOn = time.Now()
	if err := l.saveLocked(state); err != nil {
		return "", fmt.Errorf("failed to persist refreshed token: %w", err)
	}
	return newTok.AccessToken, nil
}

// OAuthManager is implemented by MCP handles for connectors with auth "oauth".
// Rill Developer's local server uses it to run the browser-based authorization flow and persist the result locally.
type OAuthManager interface {
	// OAuthEnabled reports whether the connector uses per-user OAuth.
	OAuthEnabled() bool
	// OAuthStatus reports whether a token is persisted locally and when it expires (zero if it does not expire).
	OAuthStatus(ctx context.Context) (connected bool, expiresOn time.Time, err error)
	// BeginOAuth discovers the server's authorization server (registering a client if needed) and returns the URL to send the browser to.
	// The returned verifier must be passed to CompleteOAuth.
	BeginOAuth(ctx context.Context, redirectURL, state string) (authURL, verifier string, err error)
	// CompleteOAuth exchanges the authorization code for a token and persists it locally.
	// iss is the "iss" parameter of the authorization response, if any (RFC 9207).
	CompleteOAuth(ctx context.Context, code, verifier, iss string) error
}

var _ OAuthManager = &handle{}

func (h *handle) OAuthEnabled() bool {
	return h.config.Auth == AuthOAuth
}

func (h *handle) OAuthStatus(ctx context.Context) (bool, time.Time, error) {
	if !h.OAuthEnabled() {
		return false, time.Time{}, ErrOAuthNotConfigured
	}
	state, err := h.oauth.load()
	if err != nil {
		return false, time.Time{}, err
	}
	if state == nil || state.Token == nil || state.Token.AccessToken == "" {
		return false, time.Time{}, nil
	}
	if !state.Token.Expiry.IsZero() && state.Token.Expiry.Before(time.Now()) && state.Token.RefreshToken == "" {
		return false, time.Time{}, nil
	}
	return true, state.Token.Expiry, nil
}

func (h *handle) BeginOAuth(ctx context.Context, redirectURL, state string) (string, string, error) {
	if !h.OAuthEnabled() {
		return "", "", ErrOAuthNotConfigured
	}
	existing, err := h.oauth.load()
	if err != nil {
		return "", "", err
	}
	// The loaded state is for the current configuration; its client can be reused unless Rill Developer now runs on another port.
	var client *oauthclient.Client
	if existing != nil && existing.Client.RedirectURL == redirectURL {
		client = existing.Client
	} else {
		client, err = oauthclient.Discover(ctx, &oauthclient.DiscoverOptions{
			ServerURL:    h.config.URL,
			RedirectURL:  redirectURL,
			ClientID:     h.config.OAuthClientID,
			ClientSecret: h.config.OAuthClientSecret,
			Scopes:       h.config.OAuthScopes,
			HTTPClient:   h.httpClient(),
		})
		if err != nil {
			return "", "", err
		}
		// Persist the client now so the callback can use it; the token is added when the flow completes.
		if err := h.oauth.save(&OAuthState{ConfigHash: h.oauthConfigHash, Client: client}); err != nil {
			return "", "", err
		}
	}
	verifier := oauthclient.GenerateVerifier()
	return client.AuthCodeURL(state, verifier), verifier, nil
}

func (h *handle) CompleteOAuth(ctx context.Context, code, verifier, iss string) error {
	if !h.OAuthEnabled() {
		return ErrOAuthNotConfigured
	}
	state, err := h.oauth.load()
	if err != nil {
		return err
	}
	if state == nil || state.Client == nil {
		return errors.New("no pending OAuth flow: start the flow again")
	}
	if err := state.Client.ValidateIssuer(iss); err != nil {
		return err
	}
	tok, err := state.Client.Exchange(ctx, code, verifier, h.httpClient())
	if err != nil {
		return err
	}
	h.dropSessions()
	return h.oauth.save(&OAuthState{ConfigHash: state.ConfigHash, Client: state.Client, Token: tok, TokenObtainedOn: time.Now()})
}
