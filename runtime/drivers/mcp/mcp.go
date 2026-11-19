// Package mcp implements a connector driver for external MCP (Model Context Protocol) servers.
// It lets Rill's AI agents discover and call tools on a remote server over streamable HTTP.
//
// The connector's transport properties (url, headers, timeouts) configure how to reach the server.
// Its presentation properties (display_name, ai_instructions) configure how agents are told about it.
// Who may use the server and which of its tools is decided by the connector's "security" block,
// which the parser turns into security rules on the connector spec and the AI layer evaluates per user.
//
// Credentials are either project-scoped (static headers that may use {{ .env.* }} templating)
// or user-scoped (auth: oauth, where each user authorizes the server in their browser through Rill Cloud,
// which then supplies the user's access token on every request).
// Per-user data always travels in drivers.MCPCallOptions; the handle itself is shared by all users of an instance and holds no OAuth state.
package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/mitchellh/mapstructure"
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/oauthclient"
	"github.com/rilldata/rill/runtime/pkg/safehttp"
	"github.com/rilldata/rill/runtime/storage"
	"go.uber.org/zap"
)

func init() {
	drivers.Register("mcp", driver{})
	drivers.RegisterAsConnector("mcp", driver{})
}

const (
	defaultCallTimeout    = 60 * time.Second
	defaultConnectTimeout = 10 * time.Second
	defaultMaxResultBytes = 128 * 1024
	toolsCacheTTL         = 5 * time.Minute
	maxCachedSessions     = 64
	clientName            = "rill"
	clientVersion         = "1.0.0"
)

// AuthOAuth is the value of the "auth" property for servers that require a per-user OAuth authorization.
const AuthOAuth = "oauth"

var spec = drivers.Spec{
	DisplayName: "MCP Server",
	Description: "Connect an external MCP server to make its tools available to Rill's AI agents.",
	DocsURL:     "https://docs.rilldata.com/reference/project-files/connectors#mcp-server",
	ConfigProperties: []*drivers.PropertySpec{
		{
			Key:         "url",
			Type:        drivers.StringPropertyType,
			DisplayName: "URL",
			Description: "URL of the MCP server's streamable HTTP endpoint.",
			Placeholder: "https://example.com/mcp",
			Required:    true,
		},
		{
			Key:         "auth",
			Type:        drivers.StringPropertyType,
			DisplayName: "Authentication",
			Description: "Set to \"oauth\" for servers that require each user to authorize access in their browser (Rill Cloud only).",
		},
		{
			Key:         "oauth_client_id",
			Type:        drivers.StringPropertyType,
			DisplayName: "OAuth client ID",
			Description: "Pre-registered OAuth client ID, for authorization servers without dynamic client registration.",
		},
		{
			Key:         "oauth_client_secret",
			Type:        drivers.StringPropertyType,
			DisplayName: "OAuth client secret",
			Description: "Secret of the pre-registered OAuth client.",
			Secret:      true,
		},
		{
			Key:         "display_name",
			Type:        drivers.StringPropertyType,
			DisplayName: "Display name",
			Description: "Human-friendly name for the server, shown in the UI and to AI agents.",
		},
		{
			Key:         "ai_instructions",
			Type:        drivers.StringPropertyType,
			DisplayName: "AI instructions",
			Description: "Guidance for AI agents on when and how to use the server's tools.",
		},
	},
	ImplementsMCP: true,
}

type driver struct{}

var _ drivers.Driver = driver{}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) Open(_ context.Context, connectorName, instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if instanceID == "" {
		return nil, errors.New("mcp driver can't be shared")
	}
	cfg, err := newConfigProperties(config)
	if err != nil {
		return nil, err
	}
	h := &handle{
		name:        connectorName,
		instanceID:  instanceID,
		config:      cfg,
		rawConfig:   config,
		logger:      logger.With(zap.String("connector", connectorName)),
		transport:   safehttp.NewTransport(cfg.AllowHostAccess, cfg.connectTimeout),
		userHeaders: hasUserTemplating(cfg.Headers),
		sessions:    make(map[string]*cachedSession),
		tools:       make(map[string]*cachedTools),
	}
	if cfg.Auth == AuthOAuth {
		h.oauthConfigHash = oauthConfigHash(cfg)
	}
	return h, nil
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, srcProps map[string]any, logger *zap.Logger) (bool, error) {
	return false, errors.New("not implemented")
}

func (d driver) TertiarySourceConnectors(ctx context.Context, srcProps map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, errors.New("not implemented")
}

// configProperties are the connector's properties.
// Header values may contain templating ({{ .env.* }} and {{ .user.* }}); they are resolved per request in resolveCredentials
// with the environment, variables and user attributes provided in drivers.MCPCallOptions.
type configProperties struct {
	URL               string   `mapstructure:"url"`
	Auth              string   `mapstructure:"auth"`
	OAuthClientID     string   `mapstructure:"oauth_client_id"`
	OAuthClientSecret string   `mapstructure:"oauth_client_secret"`
	OAuthScopes       []string `mapstructure:"oauth_scopes"`
	// OAuthAuthorizationParams are extra query parameters for the authorization endpoint, such as Google's access_type=offline.
	OAuthAuthorizationParams map[string]string `mapstructure:"oauth_authorization_params"`
	Headers                  map[string]string `mapstructure:"headers"`
	Timeout                  string            `mapstructure:"timeout"`
	ConnectTimeout           string            `mapstructure:"connect_timeout"`
	MaxResultBytes           int64             `mapstructure:"max_result_bytes"`
	DisplayName              string            `mapstructure:"display_name"`
	AIInstructions           string            `mapstructure:"ai_instructions"`
	AllowHostAccess          bool              `mapstructure:"allow_host_access"`

	timeout        time.Duration
	connectTimeout time.Duration
}

func newConfigProperties(config map[string]any) (*configProperties, error) {
	// WeakDecode would turn booleans into "1" or "0", which servers such as Google reject, and null into an empty string.
	// Numbers already decode correctly, so only booleans are converted here.
	if params, ok := config["oauth_authorization_params"].(map[string]any); ok {
		converted := make(map[string]any, len(params))
		for k, v := range params {
			switch v := v.(type) {
			case bool:
				converted[k] = strconv.FormatBool(v)
			case nil:
				return nil, fmt.Errorf("invalid 'oauth_authorization_params': %q has no value", k)
			default:
				converted[k] = v
			}
		}
		config = maps.Clone(config)
		config["oauth_authorization_params"] = converted
	}

	cfg := &configProperties{}
	err := mapstructure.WeakDecode(config, cfg)
	if err != nil {
		return nil, err
	}

	if cfg.URL == "" {
		return nil, errors.New("missing required property 'url'")
	}
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid 'url': %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("invalid 'url': scheme must be http or https")
	}
	if u.Host == "" {
		return nil, errors.New("invalid 'url': missing host")
	}
	if !cfg.AllowHostAccess {
		if u.Scheme != "https" {
			return nil, errors.New("invalid 'url': MCP servers must use https in hosted deployments")
		}
		if safehttp.IsPrivateHost(u.Hostname()) {
			return nil, errors.New("invalid 'url': MCP servers on private networks cannot be reached from hosted deployments")
		}
	}

	if cfg.Auth != "" && cfg.Auth != AuthOAuth {
		return nil, fmt.Errorf("invalid 'auth': expected \"oauth\" or empty, got %q", cfg.Auth)
	}
	if len(cfg.OAuthAuthorizationParams) > 0 && cfg.Auth != AuthOAuth {
		return nil, errors.New("'oauth_authorization_params' requires 'auth: oauth'")
	}
	for k := range cfg.OAuthAuthorizationParams {
		if strings.EqualFold(k, "scope") {
			return nil, fmt.Errorf("invalid 'oauth_authorization_params': cannot set %q; use 'oauth_scopes' instead", k)
		}
		if oauthclient.IsReservedAuthorizationParam(k) {
			return nil, fmt.Errorf("invalid 'oauth_authorization_params': cannot set %q, which Rill sets itself", k)
		}
	}
	for k := range cfg.Headers {
		if strings.EqualFold(k, "authorization") && cfg.Auth == AuthOAuth {
			return nil, errors.New("an 'Authorization' header cannot be combined with 'auth: oauth'")
		}
	}

	cfg.timeout = defaultCallTimeout
	if cfg.Timeout != "" {
		cfg.timeout, err = time.ParseDuration(cfg.Timeout)
		if err != nil || cfg.timeout <= 0 {
			return nil, fmt.Errorf("invalid 'timeout': expected a duration such as \"60s\"")
		}
	}
	cfg.connectTimeout = defaultConnectTimeout
	if cfg.ConnectTimeout != "" {
		cfg.connectTimeout, err = time.ParseDuration(cfg.ConnectTimeout)
		if err != nil || cfg.connectTimeout <= 0 {
			return nil, fmt.Errorf("invalid 'connect_timeout': expected a duration such as \"10s\"")
		}
	}
	if cfg.MaxResultBytes <= 0 {
		cfg.MaxResultBytes = defaultMaxResultBytes
	}

	return cfg, nil
}

// handle is a connection to an external MCP server.
type handle struct {
	name       string
	instanceID string
	config     *configProperties
	rawConfig  map[string]any
	logger     *zap.Logger
	transport  *http.Transport
	// oauthConfigHash identifies the connector's OAuth configuration (see oauthConfigHash). It is empty without auth "oauth".
	oauthConfigHash string
	// userHeaders is true if a header template references the requesting user ({{ .user.* }}),
	// which makes the server's credentials per-user even without OAuth.
	userHeaders bool

	mu       sync.Mutex
	sessions map[string]*cachedSession
	tools    map[string]*cachedTools
}

var (
	_ drivers.Handle     = &handle{}
	_ drivers.MCPService = &handle{}
)

type cachedSession struct {
	session  *gomcp.ClientSession
	rt       *authRoundTripper
	lastUsed time.Time
}

type cachedTools struct {
	tools   []*drivers.MCPTool
	expires time.Time
}

// Ping checks that the server is reachable by connecting to it.
// It does not authenticate: a server that rejects the unauthenticated handshake with 401 or 403 is up, it just requires credentials.
// The connector reconciler separately lists the server's tools with the project's credentials.
func (h *handle) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, h.config.connectTimeout)
	defer cancel()
	rt := &authRoundTripper{base: h.transport}
	client := gomcp.NewClient(&gomcp.Implementation{Name: clientName, Version: clientVersion}, nil)
	sess, err := client.Connect(ctx, &gomcp.StreamableClientTransport{
		Endpoint:             h.config.URL,
		HTTPClient:           &http.Client{Transport: rt, CheckRedirect: safehttp.CheckSameOriginRedirect},
		MaxRetries:           1,
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		if rt.rejected() {
			return nil
		}
		return fmt.Errorf("failed to connect to MCP server: %w", err)
	}
	return sess.Close()
}

func (h *handle) Driver() string {
	return "mcp"
}

func (h *handle) Config() map[string]any {
	return h.rawConfig
}

func (h *handle) Migrate(ctx context.Context) error {
	return nil
}

func (h *handle) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

func (h *handle) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	var errs []error
	for key, s := range h.sessions {
		errs = append(errs, s.session.Close())
		delete(h.sessions, key)
	}
	h.transport.CloseIdleConnections()
	return errors.Join(errs...)
}

func (h *handle) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

func (h *handle) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

func (h *handle) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

func (h *handle) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

func (h *handle) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

func (h *handle) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return nil, false
}

func (h *handle) AsInformationSchema() (drivers.InformationSchema, bool) {
	return nil, false
}

func (h *handle) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

func (h *handle) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

func (h *handle) AsWarehouse() (drivers.Warehouse, bool) {
	return nil, false
}

func (h *handle) AsModelExecutor(instanceID string, opts *drivers.ModelExecutorOptions) (drivers.ModelExecutor, error) {
	return nil, drivers.ErrNotImplemented
}

func (h *handle) AsModelManager(instanceID string) (drivers.ModelManager, error) {
	return nil, drivers.ErrNotImplemented
}

func (h *handle) AsNotifier(properties map[string]any) (drivers.Notifier, error) {
	return nil, drivers.ErrNotNotifier
}

func (h *handle) AsMCP() (drivers.MCPService, bool) {
	return h, true
}

// OAuthConfig implements drivers.MCPService.
func (h *handle) OAuthConfig() *drivers.MCPOAuthConfig {
	if h.config.Auth != AuthOAuth {
		return nil
	}
	return &drivers.MCPOAuthConfig{
		ServerURL:           h.config.URL,
		ClientID:            h.config.OAuthClientID,
		ClientSecret:        h.config.OAuthClientSecret,
		Scopes:              h.config.OAuthScopes,
		AuthorizationParams: h.config.OAuthAuthorizationParams,
		Hash:                h.oauthConfigHash,
	}
}

// ListTools implements drivers.MCPService.
func (h *handle) ListTools(ctx context.Context, opts *drivers.MCPCallOptions) ([]*drivers.MCPTool, error) {
	var res []*drivers.MCPTool
	err := h.do(ctx, opts, func(ctx context.Context, creds *credentials) error {
		h.mu.Lock()
		cached, ok := h.tools[creds.key]
		h.mu.Unlock()
		if ok && time.Now().Before(cached.expires) {
			res = cached.tools
			return nil
		}

		ctx, cancel := context.WithTimeout(ctx, h.config.timeout)
		defer cancel()

		var tools []*drivers.MCPTool
		err := h.withSession(ctx, creds, func(sess *gomcp.ClientSession) error {
			tools = tools[:0]
			for t, err := range sess.Tools(ctx, nil) {
				if err != nil {
					return err
				}
				tools = append(tools, convertTool(t))
			}
			return nil
		})
		if err != nil {
			return err
		}

		h.mu.Lock()
		if len(h.tools) >= maxCachedSessions {
			// Drop expired entries so the cache stays bounded when many users share the handle.
			now := time.Now()
			for key, entry := range h.tools {
				if now.After(entry.expires) {
					delete(h.tools, key)
				}
			}
		}
		h.tools[creds.key] = &cachedTools{tools: tools, expires: time.Now().Add(toolsCacheTTL)}
		h.mu.Unlock()
		res = tools
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// CallTool implements drivers.MCPService.
func (h *handle) CallTool(ctx context.Context, name string, args map[string]any, opts *drivers.MCPCallOptions) (*drivers.MCPToolResult, error) {
	if args == nil {
		args = map[string]any{}
	}
	var res *drivers.MCPToolResult
	err := h.do(ctx, opts, func(ctx context.Context, creds *credentials) error {
		ctx, cancel := context.WithTimeout(ctx, h.config.timeout)
		defer cancel()

		var out *gomcp.CallToolResult
		err := h.withSession(ctx, creds, func(sess *gomcp.ClientSession) error {
			var err error
			out, err = sess.CallTool(ctx, &gomcp.CallToolParams{Name: name, Arguments: args})
			return err
		})
		if err != nil {
			return err
		}
		res = h.convertResult(out)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// do resolves the caller's credentials and runs fn with them.
func (h *handle) do(ctx context.Context, opts *drivers.MCPCallOptions, fn func(ctx context.Context, creds *credentials) error) error {
	if opts == nil {
		opts = &drivers.MCPCallOptions{}
	}
	creds, err := h.resolveCredentials(opts)
	if err != nil {
		return err
	}
	return fn(ctx, creds)
}

// credentials are the resolved per-caller request credentials.
type credentials struct {
	headers map[string]string
	token   string
	key     string
}

// resolveCredentials resolves the connector's header templates for the caller and determines the bearer token to use.
func (h *handle) resolveCredentials(opts *drivers.MCPCallOptions) (*credentials, error) {
	creds := &credentials{headers: make(map[string]string, len(h.config.Headers))}

	// Callers without a user (such as the connector reconciler) cannot use a server whose headers identify the user.
	if opts.UserAttributes == nil && h.userHeaders {
		return nil, fmt.Errorf("%w: the server's headers reference the requesting user", drivers.ErrMCPAuthorizationRequired)
	}
	attrs := opts.UserAttributes
	if attrs == nil {
		attrs = map[string]any{}
	}
	td := parser.TemplateData{
		Environment: opts.Environment,
		Variables:   opts.Variables,
		User:        attrs,
	}
	for k, v := range h.config.Headers {
		resolved, err := parser.ResolveTemplate(v, td, true)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve header %q: %w", k, err)
		}
		creds.headers[k] = resolved
	}

	if h.config.Auth == AuthOAuth {
		if opts.OAuthToken == "" {
			return nil, drivers.ErrMCPAuthorizationRequired
		}
		creds.token = opts.OAuthToken
	}

	// Key the session and tool caches by the resolved credentials, so users with different credentials never share a session.
	keys := make([]string, 0, len(creds.headers))
	for k := range creds.headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	hash := sha256.New()
	for _, k := range keys {
		hash.Write([]byte(k))
		hash.Write([]byte{0})
		hash.Write([]byte(creds.headers[k]))
		hash.Write([]byte{0})
	}
	hash.Write([]byte(creds.token))
	creds.key = hex.EncodeToString(hash.Sum(nil))

	return creds, nil
}

// withSession runs fn with a client session for the credentials, reconnecting once if the server closed or forgot the session.
// Authorization failures are reported as drivers.ErrMCPAuthorizationRequired.
func (h *handle) withSession(ctx context.Context, creds *credentials, fn func(sess *gomcp.ClientSession) error) error {
	for attempt := range 2 {
		cs, err := h.session(ctx, creds)
		if err != nil {
			if cs != nil && cs.rt.unauthorized() {
				return fmt.Errorf("%w: %w", drivers.ErrMCPAuthorizationRequired, err)
			}
			return err
		}
		err = fn(cs.session)
		if err == nil {
			return nil
		}
		if cs.rt.unauthorized() {
			h.dropSession(creds.key)
			return fmt.Errorf("%w: %w", drivers.ErrMCPAuthorizationRequired, err)
		}
		if (errors.Is(err, gomcp.ErrConnectionClosed) || errors.Is(err, gomcp.ErrSessionMissing)) && attempt == 0 {
			h.dropSession(creds.key)
			continue
		}
		return err
	}
	return errors.New("failed to reconnect to MCP server")
}

// session returns a cached client session for the credentials, connecting if necessary.
// On a connection error, the returned cachedSession (if any) can still be inspected for the response status.
func (h *handle) session(ctx context.Context, creds *credentials) (*cachedSession, error) {
	h.mu.Lock()
	if cs, ok := h.sessions[creds.key]; ok {
		cs.lastUsed = time.Now()
		h.mu.Unlock()
		return cs, nil
	}
	h.mu.Unlock()

	rt := &authRoundTripper{base: h.transport, headers: creds.headers, token: creds.token}
	client := gomcp.NewClient(&gomcp.Implementation{Name: clientName, Version: clientVersion}, nil)
	transport := &gomcp.StreamableClientTransport{
		Endpoint:             h.config.URL,
		HTTPClient:           &http.Client{Transport: rt, CheckRedirect: safehttp.CheckSameOriginRedirect},
		MaxRetries:           1,
		DisableStandaloneSSE: true,
	}
	connectCtx, cancel := context.WithTimeout(ctx, h.config.connectTimeout)
	defer cancel()
	sess, err := client.Connect(connectCtx, transport, nil)
	if err != nil {
		return &cachedSession{rt: rt}, fmt.Errorf("failed to connect to MCP server: %w", err)
	}

	cs := &cachedSession{session: sess, rt: rt, lastUsed: time.Now()}
	h.mu.Lock()
	defer h.mu.Unlock()
	if existing, ok := h.sessions[creds.key]; ok {
		// Another caller connected concurrently; keep theirs.
		_ = sess.Close()
		existing.lastUsed = time.Now()
		return existing, nil
	}
	if len(h.sessions) >= maxCachedSessions {
		var oldestKey string
		var oldest time.Time
		for k, s := range h.sessions {
			if oldestKey == "" || s.lastUsed.Before(oldest) {
				oldestKey, oldest = k, s.lastUsed
			}
		}
		_ = h.sessions[oldestKey].session.Close()
		delete(h.sessions, oldestKey)
		delete(h.tools, oldestKey)
	}
	h.sessions[creds.key] = cs
	return cs, nil
}

func (h *handle) dropSession(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if cs, ok := h.sessions[key]; ok {
		_ = cs.session.Close()
		delete(h.sessions, key)
	}
	delete(h.tools, key)
}

// oauthConfigHash identifies a connector's OAuth configuration: the server URL, the pre-registered client, the scopes and the authorization params.
// Authorizations are bound to it, so changing any of them requires the user to authorize again.
// Each field is length-prefixed so that different configurations can never produce the same input.
func oauthConfigHash(cfg *configProperties) string {
	h := sha256.New()
	for _, s := range append([]string{cfg.URL, cfg.OAuthClientID, cfg.OAuthClientSecret}, cfg.OAuthScopes...) {
		_, _ = fmt.Fprintf(h, "%d:%s;", len(s), s)
	}
	// The marker can't be confused with a length-prefixed field, so scopes and params never collide
	_, _ = h.Write([]byte("params;"))
	keys := make([]string, 0, len(cfg.OAuthAuthorizationParams))
	for k := range cfg.OAuthAuthorizationParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := cfg.OAuthAuthorizationParams[k]
		_, _ = fmt.Fprintf(h, "%d:%s;%d:%s;", len(k), k, len(v), v)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// truncateUTF8 cuts s to at most limit bytes without splitting a multi-byte character.
func truncateUTF8(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	cut := limit
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut]
}

// userTemplateRegexp matches template actions that reference the requesting user, such as {{ .user.email }}.
var userTemplateRegexp = regexp.MustCompile(`\{\{-?[^}]*\.user\b`)

// hasUserTemplating reports whether any header value references the requesting user.
func hasUserTemplating(headers map[string]string) bool {
	for _, v := range headers {
		if userTemplateRegexp.MatchString(v) {
			return true
		}
	}
	return false
}

// convertTool converts an SDK tool to the driver representation.
func convertTool(t *gomcp.Tool) *drivers.MCPTool {
	return &drivers.MCPTool{
		Name:         t.Name,
		Title:        t.Title,
		Description:  t.Description,
		InputSchema:  schemaToMap(t.InputSchema),
		OutputSchema: schemaToMap(t.OutputSchema),
	}
}

// schemaToMap converts a JSON schema of any representation to a plain map.
func schemaToMap(v any) map[string]any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	return m
}

// convertResult flattens an SDK tool result into text, applying the connector's result size limit.
func (h *handle) convertResult(res *gomcp.CallToolResult) *drivers.MCPToolResult {
	out := &drivers.MCPToolResult{IsError: res.IsError}

	var parts []string
	for _, c := range res.Content {
		switch c := c.(type) {
		case *gomcp.TextContent:
			parts = append(parts, c.Text)
		case *gomcp.ImageContent:
			parts = append(parts, fmt.Sprintf("[image omitted: %s, %d bytes]", c.MIMEType, len(c.Data)))
		case *gomcp.AudioContent:
			parts = append(parts, fmt.Sprintf("[audio omitted: %s, %d bytes]", c.MIMEType, len(c.Data)))
		case *gomcp.EmbeddedResource:
			if c.Resource != nil && c.Resource.Text != "" {
				parts = append(parts, c.Resource.Text)
			} else if c.Resource != nil {
				parts = append(parts, fmt.Sprintf("[resource omitted: %s, %s, %d bytes]", c.Resource.URI, c.Resource.MIMEType, len(c.Resource.Blob)))
			}
		case *gomcp.ResourceLink:
			parts = append(parts, fmt.Sprintf("[resource link: %s]", c.URI))
		default:
			parts = append(parts, fmt.Sprintf("[unsupported content type %T omitted]", c))
		}
	}
	out.Content = strings.Join(parts, "\n")

	if res.StructuredContent != nil {
		data, err := json.Marshal(res.StructuredContent)
		if err == nil {
			out.StructuredContent = data
		}
	}

	limit := int(h.config.MaxResultBytes)
	if len(out.Content) > limit {
		out.Content = truncateUTF8(out.Content, limit) + fmt.Sprintf("\n[truncated: the result exceeded the connector's limit of %d bytes]", limit)
		out.Truncated = true
	}
	if len(out.StructuredContent) > limit {
		// Structured content is optional; drop it rather than emit invalid JSON.
		out.StructuredContent = nil
		out.Truncated = true
	}

	return out
}
