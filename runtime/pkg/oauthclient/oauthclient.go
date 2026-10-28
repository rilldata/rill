// Package oauthclient implements the client side of OAuth 2.1 authorization code flows against protected resources:
// discovery of the resource's authorization server (RFC 9728 and RFC 8414), dynamic client registration (RFC 7591)
// with a pre-registered fallback, PKCE, resource indicators (RFC 8707), issuer validation (RFC 9207), token refresh with rotation, and revocation (RFC 7009).
// The admin service uses it to run browser-based flows on behalf of users and to refresh the resulting tokens.
package oauthclient

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
	"golang.org/x/oauth2"
)

// ErrAuthorizationRequired is returned when a user's authorization is missing or gone for good (for example, a revoked refresh token),
// so the user must authorize again.
var ErrAuthorizationRequired = errors.New("authorization required")

// maxRefreshLeeway is the longest time before expiry that a token is refreshed.
// Callers may hold a token for the duration of a request, so the token must stay valid for a while after it is handed out.
const maxRefreshLeeway = 5 * time.Minute

// Client is the discovered OAuth configuration for a protected resource together with the client Rill uses with it.
// It is shared by all users of a connector; tokens are stored separately per user.
type Client struct {
	// ServerURL is the protected resource the client was discovered for.
	ServerURL string `json:"server_url"`
	// RedirectURL is the redirect URI the client was registered with.
	RedirectURL string `json:"redirect_url"`
	// Resource is the RFC 8707 resource indicator sent in authorization and token requests.
	Resource string `json:"resource"`
	Issuer   string `json:"issuer"`
	// IssParameterSupported is true if the authorization server always sends the RFC 9207 "iss" parameter in authorization responses.
	IssParameterSupported bool     `json:"iss_parameter_supported,omitempty"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	RevocationEndpoint    string   `json:"revocation_endpoint,omitempty"`
	ClientID              string   `json:"client_id"`
	ClientSecret          string   `json:"client_secret,omitempty"`
	ClientAuthStyle       string   `json:"client_auth_style,omitempty"` // "basic", "post" or empty for auto-detect
	Scopes                []string `json:"scopes,omitempty"`
	// AuthorizationParams are extra query parameters for the authorization endpoint, for servers that need non-standard ones
	// (e.g. Google's access_type=offline, without which it issues no refresh token).
	AuthorizationParams map[string]string `json:"authorization_params,omitempty"`
}

// DiscoverOptions are the inputs for Discover.
type DiscoverOptions struct {
	ServerURL   string
	RedirectURL string
	// ClientID and ClientSecret are pre-registered client credentials, used when the authorization server does not support dynamic client registration.
	ClientID     string
	ClientSecret string
	// Scopes to request. Defaults to all scopes the server advertises.
	Scopes []string
	// AuthorizationParams are extra query parameters for the authorization endpoint (see Client.AuthorizationParams).
	AuthorizationParams map[string]string
	// HTTPClient is used for all requests; see safehttp.NewHTTPClient.
	HTTPClient *http.Client
}

// Discover discovers the authorization server of a protected resource and registers a client with it (or uses the pre-registered one).
func Discover(ctx context.Context, opts *DiscoverOptions) (*Client, error) {
	hc := opts.HTTPClient

	// The resource indicator defaults to the canonical server URL (RFC 8707 section 2).
	resource := canonicalResource(opts.ServerURL)

	// Find the authorization server through the protected resource metadata at its well-known URLs.
	// The metadata must declare the resource its URL was derived from (RFC 9728 section 3.3), which then becomes the resource indicator.
	var issuer string
	var scopes []string
	var prmErrs []error
	for _, c := range protectedResourceMetadataCandidates(resource) {
		prm, err := oauthex.GetProtectedResourceMetadata(ctx, c.url, c.resource, hc)
		if err != nil {
			prmErrs = append(prmErrs, err)
			continue
		}
		if len(prm.AuthorizationServers) == 0 {
			prmErrs = append(prmErrs, fmt.Errorf("%s does not list any authorization servers", c.url))
			continue
		}
		resource = c.resource
		issuer = prm.AuthorizationServers[0]
		scopes = prm.ScopesSupported
		break
	}
	if issuer == "" {
		// Servers that predate protected resource metadata (MCP specification 2025-03-26, e.g. Atlassian's) serve
		// their authorization server metadata on their own origin.
		u, err := url.Parse(resource)
		if err != nil {
			return nil, fmt.Errorf("invalid server url: %w", err)
		}
		issuer = u.Scheme + "://" + u.Host
	}

	meta, err := auth.GetAuthServerMetadata(ctx, issuer, hc)
	if err == nil && meta == nil {
		err = errors.New("no metadata found at its well-known URLs")
	}
	if err != nil {
		if len(prmErrs) > 0 {
			return nil, fmt.Errorf("the server advertises no usable OAuth protected resource metadata (%w), and its origin %q serves no authorization server metadata: %w", errors.Join(prmErrs...), issuer, err)
		}
		return nil, fmt.Errorf("failed to get authorization server metadata for %q: %w", issuer, err)
	}
	if meta.AuthorizationEndpoint == "" || meta.TokenEndpoint == "" {
		return nil, fmt.Errorf("authorization server %q does not advertise authorization and token endpoints", issuer)
	}
	if len(meta.CodeChallengeMethodsSupported) > 0 && !slices.Contains(meta.CodeChallengeMethodsSupported, "S256") {
		return nil, fmt.Errorf("authorization server %q does not support PKCE with S256, which OAuth 2.1 requires", issuer)
	}

	if len(opts.Scopes) > 0 {
		// Configured scopes are requested exactly as they are, so a client that may not use offline_access can opt out by listing its scopes.
		scopes = opts.Scopes
	} else if len(scopes) > 0 && slices.Contains(meta.ScopesSupported, "offline_access") && !slices.Contains(scopes, "offline_access") {
		// Rill stores tokens securely and refreshes them, so it asks for a refresh token where the authorization server lets clients request one with the offline_access scope (MCP SEP-2207).
		// Servers may still decline to issue one.
		// Without scopes, no scope is sent and the server applies its defaults, which offline_access alone would replace.
		scopes = append(scopes, "offline_access")
	}

	c := &Client{
		ServerURL:             opts.ServerURL,
		RedirectURL:           opts.RedirectURL,
		Resource:              resource,
		Issuer:                meta.Issuer,
		IssParameterSupported: meta.AuthorizationResponseIssParameterSupported,
		AuthorizationEndpoint: meta.AuthorizationEndpoint,
		TokenEndpoint:         meta.TokenEndpoint,
		RevocationEndpoint:    meta.RevocationEndpoint,
		Scopes:                scopes,
		AuthorizationParams:   opts.AuthorizationParams,
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

// TokenExpiring reports whether a token should be refreshed before use: it has an expiry, and that expiry is within the refresh leeway.
// obtainedOn is when the token was issued or last refreshed. The leeway is maxRefreshLeeway, capped at half the token's lifetime,
// so short-lived tokens are not refreshed on every use.
func TokenExpiring(tok *oauth2.Token, obtainedOn time.Time) bool {
	if tok.Expiry.IsZero() {
		return false
	}
	leeway := min(maxRefreshLeeway, tok.Expiry.Sub(obtainedOn)/2)
	return time.Until(tok.Expiry) < leeway
}

// AuthCodeURL returns the URL to send the user's browser to.
func (c *Client) AuthCodeURL(state, verifier string) string {
	opts := make([]oauth2.AuthCodeOption, 0, len(c.AuthorizationParams)+2)
	for k, v := range c.AuthorizationParams {
		opts = append(opts, oauth2.SetAuthURLParam(k, v))
	}
	opts = append(opts, oauth2.S256ChallengeOption(verifier), oauth2.SetAuthURLParam("resource", c.Resource))
	return c.oauth2Config().AuthCodeURL(state, opts...)
}

// reservedAuthorizationParams are the authorization request parameters that the flow sets itself or depends on:
// "request" and "request_uri" would let a request object (RFC 9101) replace them, and "response_mode" would move the response out of the callback's query.
var reservedAuthorizationParams = []string{
	"response_type", "client_id", "redirect_uri", "scope", "state", "code_challenge", "code_challenge_method", "resource",
	"request", "request_uri", "response_mode",
}

// IsReservedAuthorizationParam reports whether an authorization request parameter is set by the flow itself, and so can't be configured.
func IsReservedAuthorizationParam(key string) bool {
	return slices.ContainsFunc(reservedAuthorizationParams, func(r string) bool { return strings.EqualFold(r, key) })
}

// ValidateIssuer checks the "iss" parameter of an authorization response (RFC 9207), which defends against mix-up attacks
// when Rill talks to several authorization servers through the same redirect URI.
// It is required when the authorization server advertises support for it; otherwise it is optional, but must match when present.
func (c *Client) ValidateIssuer(iss string) error {
	if iss == "" {
		if c.IssParameterSupported {
			return errors.New("the authorization server did not identify itself in the authorization response")
		}
		return nil
	}
	if iss != c.Issuer {
		return fmt.Errorf("the authorization response came from %q, not the expected authorization server %q", iss, c.Issuer)
	}
	return nil
}

// Exchange exchanges an authorization code for a token.
func (c *Client) Exchange(ctx context.Context, code, verifier string, hc *http.Client) (*oauth2.Token, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, hc)
	tok, err := c.oauth2Config().Exchange(ctx, code, oauth2.VerifierOption(verifier), oauth2.SetAuthURLParam("resource", c.Resource))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}
	return tok, nil
}

// Refresh obtains a new token using the refresh token, sending the resource indicator (RFC 8707).
// If the server rotates the refresh token, the new one is returned; otherwise the old one is kept.
func (c *Client) Refresh(ctx context.Context, tok *oauth2.Token, hc *http.Client) (*oauth2.Token, error) {
	if tok.RefreshToken == "" {
		return nil, fmt.Errorf("%w: the token cannot be refreshed", ErrAuthorizationRequired)
	}

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {tok.RefreshToken},
		"resource":      {c.Resource},
	}
	// The scope is omitted, so the new token has the scopes of the original grant (RFC 6749 section 6),
	// which may be fewer than the scopes requested.

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
		return nil, fmt.Errorf("%w: %w", ErrAuthorizationRequired, lastErr)
	}
	return nil, fmt.Errorf("failed to refresh token: %w", lastErr)
}

// Revoke revokes the token at the authorization server, if it supports revocation.
// Revoking the refresh token also ends the access tokens of its grant (RFC 7009 section 2.1).
func (c *Client) Revoke(ctx context.Context, tok *oauth2.Token, hc *http.Client) error {
	if c.RevocationEndpoint == "" {
		return nil
	}
	token, hint := tok.RefreshToken, "refresh_token"
	if token == "" {
		token, hint = tok.AccessToken, "access_token"
	}
	form := url.Values{"token": {token}, "token_type_hint": {hint}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.RevocationEndpoint, http.NoBody)
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

func (c *Client) oauth2Config() *oauth2.Config {
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

func (c *Client) postTokenEndpoint(ctx context.Context, hc *http.Client, form url.Values, style string) (*tokenResponse, error) {
	form = cloneValues(form)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.TokenEndpoint, http.NoBody)
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

// applyClientAuth adds client authentication to a token request.
func (c *Client) applyClientAuth(req *http.Request, form url.Values, style string) {
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

// GenerateState returns a random, URL-safe state parameter.
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateVerifier returns a random PKCE code verifier.
func GenerateVerifier() string {
	return oauth2.GenerateVerifier()
}

// protectedResourceMetadataCandidate is a well-known protected resource metadata URL and the resource its metadata must declare.
type protectedResourceMetadataCandidate struct {
	url      string
	resource string
}

// protectedResourceMetadataCandidates returns the well-known metadata URLs for a server URL, most specific first (RFC 9728 section 3).
// Metadata at the root URL describes the origin, but servers commonly declare their full URL there, so both are accepted.
func protectedResourceMetadataCandidates(serverURL string) []protectedResourceMetadataCandidate {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil
	}
	origin := u.Scheme + "://" + u.Host
	root := origin + "/.well-known/oauth-protected-resource"
	p := strings.Trim(u.Path, "/")
	if p == "" {
		return []protectedResourceMetadataCandidate{{url: root, resource: serverURL}}
	}
	return []protectedResourceMetadataCandidate{
		{url: root + "/" + p, resource: serverURL},
		{url: root, resource: serverURL},
		{url: root, resource: origin},
	}
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
