package mcp

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
	"github.com/rilldata/rill/runtime/drivers"
	"golang.org/x/oauth2"
)

// This file implements the OAuth 2.1 client side of the MCP authorization spec:
// discovery of the server's authorization server (RFC 9728 and RFC 8414), dynamic client registration (RFC 7591)
// with a pre-registered fallback, PKCE, resource indicators (RFC 8707), and token refresh with rotation.
// It is shared by the driver (which consumes tokens), Rill Developer's local server and the admin service
// (which run the browser-based flow on behalf of a user).

// ErrOAuthNotConfigured is returned when an OAuth operation is attempted on a server without auth "oauth".
var ErrOAuthNotConfigured = errors.New("the connector does not use OAuth (set 'auth: oauth' to enable it)")

// oauthStateFilename is the file in the connector's data dir that holds locally persisted OAuth state (Rill Developer).
const oauthStateFilename = "oauth.json"

// tokenRefreshLeeway is how long before expiry a token is refreshed.
const tokenRefreshLeeway = 60 * time.Second

// OAuthClient is the discovered OAuth configuration for an MCP server together with the client Rill uses with it.
// It is shared by all users of a connector; tokens are stored separately per user.
type OAuthClient struct {
	// ServerURL is the MCP server the client was discovered for; the client is discarded if the URL changes.
	ServerURL string `json:"server_url"`
	// RedirectURL is the redirect URI the client was registered with.
	RedirectURL string `json:"redirect_url"`
	// Resource is the RFC 8707 resource indicator sent in authorization and token requests.
	Resource              string   `json:"resource"`
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	RevocationEndpoint    string   `json:"revocation_endpoint,omitempty"`
	ClientID              string   `json:"client_id"`
	ClientSecret          string   `json:"client_secret,omitempty"`
	ClientAuthStyle       string   `json:"client_auth_style,omitempty"` // "basic", "post" or empty for auto-detect
	Scopes                []string `json:"scopes,omitempty"`
}

// DiscoverOAuthClientOptions are the inputs for DiscoverOAuthClient.
type DiscoverOAuthClientOptions struct {
	ServerURL   string
	RedirectURL string
	// ClientID and ClientSecret are pre-registered client credentials, used when the authorization server does not support dynamic client registration.
	ClientID     string
	ClientSecret string
	// Scopes to request. Defaults to all scopes the server advertises.
	Scopes     []string
	HTTPClient *http.Client
}

// DiscoverOAuthClient discovers the authorization server of an MCP server and registers a client with it (or uses the pre-registered one).
func DiscoverOAuthClient(ctx context.Context, opts *DiscoverOAuthClientOptions) (*OAuthClient, error) {
	hc := opts.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}

	// Find the protected resource metadata, preferably from the server's own 401 challenge.
	var metadataURLs []string
	challenge, err := probeServer(ctx, opts.ServerURL, hc)
	if err != nil {
		return nil, err
	}
	if challenge != nil && challenge.resourceMetadataURL != "" {
		metadataURLs = append(metadataURLs, challenge.resourceMetadataURL)
	}
	metadataURLs = append(metadataURLs, protectedResourceMetadataURLs(opts.ServerURL)...)

	var prm *oauthex.ProtectedResourceMetadata
	var prmErrs []error
	for _, u := range metadataURLs {
		prm, err = fetchProtectedResourceMetadata(ctx, u, opts.ServerURL, hc)
		if err == nil {
			break
		}
		prmErrs = append(prmErrs, fmt.Errorf("%s: %w", u, err))
	}
	if prm == nil {
		return nil, fmt.Errorf("the MCP server does not advertise OAuth protected resource metadata: %w", errors.Join(prmErrs...))
	}

	// Discover the authorization server.
	issuer := prm.AuthorizationServers[0]
	meta, err := auth.GetAuthServerMetadata(ctx, issuer, hc)
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization server metadata for %q: %w", issuer, err)
	}
	if meta.AuthorizationEndpoint == "" || meta.TokenEndpoint == "" {
		return nil, fmt.Errorf("authorization server %q does not advertise authorization and token endpoints", issuer)
	}
	if len(meta.CodeChallengeMethodsSupported) > 0 && !slices.Contains(meta.CodeChallengeMethodsSupported, "S256") {
		return nil, fmt.Errorf("authorization server %q does not support PKCE with S256, which the MCP specification requires", issuer)
	}

	scopes := opts.Scopes
	if len(scopes) == 0 && challenge != nil && challenge.scope != "" {
		scopes = strings.Fields(challenge.scope)
	}
	if len(scopes) == 0 {
		scopes = prm.ScopesSupported
	}

	resource := prm.Resource
	if resource == "" {
		resource = canonicalResource(opts.ServerURL)
	}

	c := &OAuthClient{
		ServerURL:             opts.ServerURL,
		RedirectURL:           opts.RedirectURL,
		Resource:              resource,
		Issuer:                meta.Issuer,
		AuthorizationEndpoint: meta.AuthorizationEndpoint,
		TokenEndpoint:         meta.TokenEndpoint,
		RevocationEndpoint:    meta.RevocationEndpoint,
		Scopes:                scopes,
	}

	// Use the pre-registered client if provided, otherwise register one dynamically.
	if opts.ClientID != "" {
		c.ClientID = opts.ClientID
		c.ClientSecret = opts.ClientSecret
		c.ClientAuthStyle = clientAuthStyle(meta.TokenEndpointAuthMethodsSupported)
		return c, nil
	}
	if meta.RegistrationEndpoint == "" {
		return nil, fmt.Errorf("authorization server %q does not support dynamic client registration; set 'oauth_client_id' and 'oauth_client_secret' on the connector", issuer)
	}
	reg, err := oauthex.RegisterClient(ctx, meta.RegistrationEndpoint, &oauthex.ClientRegistrationMetadata{
		RedirectURIs:            []string{opts.RedirectURL},
		TokenEndpointAuthMethod: "none",
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		ClientName:              "Rill",
		ClientURI:               "https://www.rilldata.com",
		Scope:                   strings.Join(scopes, " "),
	}, hc)
	if err != nil {
		return nil, fmt.Errorf("failed to register an OAuth client with %q: %w", issuer, err)
	}
	c.ClientID = reg.ClientID
	c.ClientSecret = reg.ClientSecret
	if c.ClientSecret != "" {
		c.ClientAuthStyle = clientAuthStyle(meta.TokenEndpointAuthMethodsSupported)
	}
	return c, nil
}

// clientAuthStyle picks how to send client credentials to the token endpoint based on what the server supports.
func clientAuthStyle(supported []string) string {
	hasPost := slices.Contains(supported, "client_secret_post")
	hasBasic := slices.Contains(supported, "client_secret_basic")
	switch {
	case hasPost && !hasBasic:
		return "post"
	case hasBasic && !hasPost:
		return "basic"
	default:
		return ""
	}
}

// Matches reports whether the client is still valid for the given server URL and redirect URL.
func (c *OAuthClient) Matches(serverURL, redirectURL string) bool {
	return c != nil && c.ServerURL == serverURL && c.RedirectURL == redirectURL
}

// AuthCodeURL returns the URL to send the user's browser to.
func (c *OAuthClient) AuthCodeURL(state, verifier string) string {
	return c.oauth2Config().AuthCodeURL(state, oauth2.S256ChallengeOption(verifier), oauth2.SetAuthURLParam("resource", c.Resource))
}

// Exchange exchanges an authorization code for a token.
func (c *OAuthClient) Exchange(ctx context.Context, code, verifier string, hc *http.Client) (*oauth2.Token, error) {
	ctx = withHTTPClient(ctx, hc)
	tok, err := c.oauth2Config().Exchange(ctx, code, oauth2.VerifierOption(verifier), oauth2.SetAuthURLParam("resource", c.Resource))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}
	return tok, nil
}

// Refresh obtains a new token using the refresh token, sending the resource indicator as the MCP specification requires.
// If the server rotates the refresh token, the new one is returned; otherwise the old one is kept.
func (c *OAuthClient) Refresh(ctx context.Context, tok *oauth2.Token, hc *http.Client) (*oauth2.Token, error) {
	if tok == nil || tok.RefreshToken == "" {
		return nil, fmt.Errorf("%w: the token cannot be refreshed", drivers.ErrMCPAuthorizationRequired)
	}
	if hc == nil {
		hc = http.DefaultClient
	}

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {tok.RefreshToken},
		"resource":      {c.Resource},
	}
	if len(c.Scopes) > 0 {
		form.Set("scope", strings.Join(c.Scopes, " "))
	}

	styles := []string{c.ClientAuthStyle}
	if c.ClientAuthStyle == "" && c.ClientSecret != "" {
		styles = []string{"basic", "post"}
	}
	var lastErr error
	for _, style := range styles {
		res, err := c.postTokenEndpoint(ctx, hc, form, style)
		if err == nil {
			out := &oauth2.Token{
				AccessToken:  res.AccessToken,
				TokenType:    res.TokenType,
				RefreshToken: res.RefreshToken,
			}
			if out.RefreshToken == "" {
				out.RefreshToken = tok.RefreshToken
			}
			if res.ExpiresIn > 0 {
				out.Expiry = time.Now().Add(time.Duration(res.ExpiresIn) * time.Second)
			}
			return out, nil
		}
		lastErr = err
		var terr *tokenEndpointError
		if !errors.As(err, &terr) || terr.code != "invalid_client" {
			break
		}
	}
	var terr *tokenEndpointError
	if errors.As(lastErr, &terr) && (terr.code == "invalid_grant" || terr.status == http.StatusUnauthorized) {
		return nil, fmt.Errorf("%w: %w", drivers.ErrMCPAuthorizationRequired, lastErr)
	}
	return nil, fmt.Errorf("failed to refresh token: %w", lastErr)
}

// Revoke revokes the token at the authorization server, if it supports revocation. It is best-effort.
func (c *OAuthClient) Revoke(ctx context.Context, tok *oauth2.Token, hc *http.Client) error {
	if c.RevocationEndpoint == "" || tok == nil {
		return nil
	}
	if hc == nil {
		hc = http.DefaultClient
	}
	token, hint := tok.RefreshToken, "refresh_token"
	if token == "" {
		token, hint = tok.AccessToken, "access_token"
	}
	form := url.Values{"token": {token}, "token_type_hint": {hint}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.RevocationEndpoint, nil)
	if err != nil {
		return err
	}
	// Client authentication may add form fields, so the body is set only after it is applied.
	c.applyClientAuth(req, form, "")
	setFormBody(req, form)
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("revocation endpoint returned status %d", resp.StatusCode)
	}
	return nil
}

func (c *OAuthClient) oauth2Config() *oauth2.Config {
	style := oauth2.AuthStyleAutoDetect
	switch c.ClientAuthStyle {
	case "basic":
		style = oauth2.AuthStyleInHeader
	case "post":
		style = oauth2.AuthStyleInParams
	}
	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   c.AuthorizationEndpoint,
			TokenURL:  c.TokenEndpoint,
			AuthStyle: style,
		},
		RedirectURL: c.RedirectURL,
		Scopes:      c.Scopes,
	}
}

// tokenResponse is the JSON body returned by a token endpoint.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
}

// tokenEndpointError is an OAuth error response from a token endpoint.
type tokenEndpointError struct {
	status      int
	code        string
	description string
}

func (e *tokenEndpointError) Error() string {
	if e.code == "" {
		return fmt.Sprintf("token endpoint returned status %d", e.status)
	}
	if e.description == "" {
		return fmt.Sprintf("token endpoint returned %q (status %d)", e.code, e.status)
	}
	return fmt.Sprintf("token endpoint returned %q: %s (status %d)", e.code, e.description, e.status)
}

func (c *OAuthClient) postTokenEndpoint(ctx context.Context, hc *http.Client, form url.Values, style string) (*tokenResponse, error) {
	form = cloneValues(form)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.TokenEndpoint, nil)
	if err != nil {
		return nil, err
	}
	c.applyClientAuth(req, form, style)
	setFormBody(req, form)
	req.Header.Set("Accept", "application/json")

	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		terr := &tokenEndpointError{status: resp.StatusCode}
		var body struct {
			Error       string `json:"error"`
			Description string `json:"error_description"`
		}
		if json.Unmarshal(data, &body) == nil {
			terr.code = body.Error
			terr.description = body.Description
		}
		return nil, terr
	}
	res := &tokenResponse{}
	if err := json.Unmarshal(data, res); err != nil {
		return nil, fmt.Errorf("invalid token endpoint response: %w", err)
	}
	if res.AccessToken == "" {
		return nil, errors.New("token endpoint response did not include an access token")
	}
	return res, nil
}

// applyClientAuth adds client authentication to a token or revocation request.
func (c *OAuthClient) applyClientAuth(req *http.Request, form url.Values, style string) {
	if style == "" {
		style = c.ClientAuthStyle
	}
	if c.ClientSecret == "" {
		form.Set("client_id", c.ClientID)
		return
	}
	switch style {
	case "post":
		form.Set("client_id", c.ClientID)
		form.Set("client_secret", c.ClientSecret)
	default:
		req.SetBasicAuth(url.QueryEscape(c.ClientID), url.QueryEscape(c.ClientSecret))
	}
}

// setFormBody sets a form-encoded body on a request, keeping the content length and GetBody consistent with it.
func setFormBody(req *http.Request, form url.Values) {
	body := form.Encode()
	req.Body = io.NopCloser(strings.NewReader(body))
	req.ContentLength = int64(len(body))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(body)), nil
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
}

func cloneValues(v url.Values) url.Values {
	out := make(url.Values, len(v))
	for k, vals := range v {
		out[k] = slices.Clone(vals)
	}
	return out
}

// withHTTPClient makes golang.org/x/oauth2 use the given client.
func withHTTPClient(ctx context.Context, hc *http.Client) context.Context {
	if hc == nil {
		return ctx
	}
	return context.WithValue(ctx, oauth2.HTTPClient, hc)
}

// GenerateOAuthState returns a random, URL-safe state parameter.
func GenerateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateOAuthVerifier returns a random PKCE code verifier.
func GenerateOAuthVerifier() string {
	return oauth2.GenerateVerifier()
}

// authChallenge is what an unauthenticated request to an MCP server revealed about its authorization requirements.
type authChallenge struct {
	resourceMetadataURL string
	scope               string
}

// probeServer sends an unauthenticated initialize request to the server.
// It returns an error if the server cannot be reached at all, and the parsed 401 challenge (if any) otherwise.
func probeServer(ctx context.Context, serverURL string, hc *http.Client) (*authChallenge, error) {
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"rill","version":"1.0.0"}}}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, serverURL, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", "2025-06-18")
	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach the MCP server: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))

	if resp.StatusCode != http.StatusUnauthorized {
		return nil, nil
	}
	challenges, err := oauthex.ParseWWWAuthenticate(resp.Header.Values("WWW-Authenticate"))
	if err != nil {
		return &authChallenge{}, nil
	}
	for _, ch := range challenges {
		if ch.Scheme != "bearer" {
			continue
		}
		return &authChallenge{
			resourceMetadataURL: ch.Params["resource_metadata"],
			scope:               ch.Params["scope"],
		}, nil
	}
	return &authChallenge{}, nil
}

// protectedResourceMetadataURLs returns the well-known metadata URLs for a server URL, most specific first (RFC 9728 section 3).
func protectedResourceMetadataURLs(serverURL string) []string {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil
	}
	origin := u.Scheme + "://" + u.Host
	var res []string
	if p := strings.Trim(u.Path, "/"); p != "" {
		res = append(res, origin+"/.well-known/oauth-protected-resource/"+p)
	}
	res = append(res, origin+"/.well-known/oauth-protected-resource")
	return res
}

// fetchProtectedResourceMetadata fetches and validates protected resource metadata.
// It accepts a declared resource that is the server's origin (as some servers do) rather than requiring an exact match.
func fetchProtectedResourceMetadata(ctx context.Context, metadataURL, serverURL string, hc *http.Client) (*oauthex.ProtectedResourceMetadata, error) {
	mu, err := url.Parse(metadataURL)
	if err != nil {
		return nil, err
	}
	su, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	if !sameOrigin(mu, su) {
		return nil, fmt.Errorf("metadata URL must be on the same origin as the server")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	prm := &oauthex.ProtectedResourceMetadata{}
	if err := json.Unmarshal(data, prm); err != nil {
		return nil, fmt.Errorf("invalid metadata: %w", err)
	}
	if len(prm.AuthorizationServers) == 0 {
		return nil, errors.New("metadata does not list any authorization servers")
	}
	if prm.Resource != "" {
		ru, err := url.Parse(prm.Resource)
		if err != nil || !sameOrigin(ru, su) {
			return nil, fmt.Errorf("metadata declares resource %q, which is not on the server's origin", prm.Resource)
		}
	}
	for _, as := range prm.AuthorizationServers {
		au, err := url.Parse(as)
		if err != nil || (au.Scheme != "https" && !isPrivateHost(au.Hostname())) {
			return nil, fmt.Errorf("metadata declares an insecure authorization server %q", as)
		}
	}
	return prm, nil
}

// canonicalResource returns the canonical URI of the server for use as a resource indicator (RFC 8707 section 2).
func canonicalResource(serverURL string) string {
	u, err := url.Parse(serverURL)
	if err != nil {
		return serverURL
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	u.Fragment = ""
	u.Path = strings.TrimSuffix(u.Path, "/")
	return u.String()
}

// OAuthState is the OAuth state Rill Developer persists for a connector: the client and the local user's token.
type OAuthState struct {
	Client *OAuthClient  `json:"client"`
	Token  *oauth2.Token `json:"token,omitempty"`
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
	// Discard state established for a different server URL.
	if state.Client == nil || state.Client.ServerURL != l.h.config.URL {
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
	expiring := !tok.Expiry.IsZero() && time.Until(tok.Expiry) < tokenRefreshLeeway
	if !forceRefresh && !expiring {
		return tok.AccessToken, nil
	}
	if tok.RefreshToken == "" {
		if expiring {
			return "", fmt.Errorf("%w: the token has expired", drivers.ErrMCPAuthorizationRequired)
		}
		return tok.AccessToken, nil
	}
	newTok, err := state.Client.Refresh(ctx, tok, l.h.httpClient())
	if err != nil {
		return "", err
	}
	state.Token = newTok
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
	CompleteOAuth(ctx context.Context, code, verifier string) error
	// OAuthClientOptions returns the discovery inputs configured on the connector (pre-registered client and scopes).
	OAuthClientOptions() *DiscoverOAuthClientOptions
}

var _ OAuthManager = &handle{}

func (h *handle) OAuthEnabled() bool {
	return h.config.Auth == AuthOAuth
}

func (h *handle) OAuthClientOptions() *DiscoverOAuthClientOptions {
	return &DiscoverOAuthClientOptions{
		ServerURL:    h.config.URL,
		ClientID:     h.config.OAuthClientID,
		ClientSecret: h.config.OAuthClientSecret,
		Scopes:       h.config.OAuthScopes,
		HTTPClient:   h.httpClient(),
	}
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
	var client *OAuthClient
	if existing != nil && existing.Client.Matches(h.config.URL, redirectURL) && (h.config.OAuthClientID == "" || existing.Client.ClientID == h.config.OAuthClientID) {
		client = existing.Client
	} else {
		opts := h.OAuthClientOptions()
		opts.RedirectURL = redirectURL
		client, err = DiscoverOAuthClient(ctx, opts)
		if err != nil {
			return "", "", err
		}
		// Persist the client now so the callback can use it; the token is added when the flow completes.
		if err := h.oauth.save(&OAuthState{Client: client}); err != nil {
			return "", "", err
		}
	}
	verifier := GenerateOAuthVerifier()
	return client.AuthCodeURL(state, verifier), verifier, nil
}

func (h *handle) CompleteOAuth(ctx context.Context, code, verifier string) error {
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
	tok, err := state.Client.Exchange(ctx, code, verifier, h.httpClient())
	if err != nil {
		return err
	}
	h.dropSessions()
	return h.oauth.save(&OAuthState{Client: state.Client, Token: tok})
}

// NewHTTPClient returns an HTTP client suitable for OAuth discovery and token requests from outside the driver
// (the admin service and Rill Developer's local server). Unless allowPrivate is set, it refuses private network addresses.
func NewHTTPClient(allowPrivate bool, timeout time.Duration) *http.Client {
	return newHTTPClient(newTransport(allowPrivate, timeout), timeout)
}

// OAuthCallbackPage renders the page shown at the end of a browser-based authorization flow.
// It notifies the window that opened the popup (the chat) and closes itself; if it was not opened as a popup,
// it offers a link back to redirectURL. errMsg is empty on success.
func OAuthCallbackPage(connector, displayName, errMsg, redirectURL string) string {
	payload, _ := json.Marshal(map[string]any{
		"type":      "rill-mcp-oauth",
		"connector": connector,
		"status":    map[bool]string{true: "error", false: "connected"}[errMsg != ""],
		"error":     errMsg,
	})
	var body string
	if errMsg != "" {
		body = fmt.Sprintf("<h1>Could not connect %s</h1><p>%s</p>", html.EscapeString(displayName), html.EscapeString(errMsg))
	} else {
		body = fmt.Sprintf("<h1>Connected %s</h1><p>You can close this window and return to the chat.</p>", html.EscapeString(displayName))
	}
	link := ""
	if isLinkableRedirect(redirectURL) {
		link = fmt.Sprintf(`<p><a href="%s">Back to Rill</a></p>`, html.EscapeString(redirectURL))
	}
	return fmt.Sprintf(`<!doctype html>
<html><head><meta charset="utf-8"><title>Rill</title>
<style>body{font-family:system-ui,sans-serif;margin:48px auto;max-width:480px;color:#1f2937}h1{font-size:20px}a{color:#2563eb}</style>
</head><body>%s%s
<script>
(function(){
  var msg=%s;
  try{ if(window.opener){ window.opener.postMessage(msg,"*"); } }catch(e){}
  if(window.opener){ setTimeout(function(){ window.close(); }, 500); }
})();
</script></body></html>`, body, link, string(payload))
}

// isLinkableRedirect reports whether a return URL may be rendered as a link on the callback page:
// a relative path or an absolute http(s) URL. The callers validate the host; this guards against javascript: and data: URLs.
func isLinkableRedirect(redirectURL string) bool {
	if redirectURL == "" {
		return false
	}
	u, err := url.Parse(redirectURL)
	if err != nil {
		return false
	}
	if u.Scheme == "" && u.Host == "" {
		return strings.HasPrefix(u.Path, "/") && !strings.HasPrefix(u.Path, "//")
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}
