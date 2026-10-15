// Package mcp implements a connector driver for external MCP (Model Context Protocol) servers.
// It lets Rill's AI agents discover and call tools on a remote server over streamable HTTP.
//
// The connector's transport properties (url, headers, timeouts) configure how to reach the server.
// Its presentation properties (display_name, ai_instructions) configure how agents are told about it.
// Who may use the server and which of its tools is decided by the connector's "security" block,
// which the parser turns into security rules on the connector spec and the AI layer evaluates per user.
//
// Credentials are project-scoped: static headers that may use {{ .env.* }} templating, resolved by the runtime before Open.
// Every user of the server therefore acts with the same identity on the server.
package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/mitchellh/mapstructure"
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
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
	clientName            = "rill"
	clientVersion         = "1.0.0"
)

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
	return &handle{
		config:    cfg,
		rawConfig: config,
		logger:    logger.With(zap.String("connector", connectorName)),
		transport: newTransport(cfg.AllowHostAccess, cfg.connectTimeout),
		sessions:  make(map[string]*cachedSession),
	}, nil
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, srcProps map[string]any, logger *zap.Logger) (bool, error) {
	return false, errors.New("not implemented")
}

func (d driver) TertiarySourceConnectors(ctx context.Context, srcProps map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, errors.New("not implemented")
}

// configProperties are the connector's properties.
// Templating in header values ({{ .env.* }}) has already been resolved by the runtime.
type configProperties struct {
	URL             string            `mapstructure:"url"`
	Headers         map[string]string `mapstructure:"headers"`
	Timeout         string            `mapstructure:"timeout"`
	ConnectTimeout  string            `mapstructure:"connect_timeout"`
	MaxResultBytes  int64             `mapstructure:"max_result_bytes"`
	DisplayName     string            `mapstructure:"display_name"`
	AIInstructions  string            `mapstructure:"ai_instructions"`
	AllowHostAccess bool              `mapstructure:"allow_host_access"`

	timeout        time.Duration
	connectTimeout time.Duration
}

func newConfigProperties(config map[string]any) (*configProperties, error) {
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
		if isPrivateHost(u.Hostname()) {
			return nil, errors.New("invalid 'url': MCP servers on private networks cannot be reached from hosted deployments")
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
// It keeps a client session per set of credentials, so callers with different credentials never share a session.
type handle struct {
	config    *configProperties
	rawConfig map[string]any
	logger    *zap.Logger
	transport *http.Transport

	mu       sync.Mutex
	sessions map[string]*cachedSession // keyed by credentials.key
}

var (
	_ drivers.Handle     = &handle{}
	_ drivers.MCPService = &handle{}
)

// cachedSession is a client session together with the tools the server exposes to its credentials.
type cachedSession struct {
	session      *gomcp.ClientSession
	tools        []*drivers.MCPTool
	toolsExpires time.Time
}

// credentials are the headers sent to the server and the key of the session that uses them.
type credentials struct {
	headers map[string]string
	key     string
}

// Ping connects to the server, which checks both that it is reachable and that it accepts the connector's credentials.
func (h *handle) Ping(ctx context.Context) error {
	_, err := h.connect(ctx, h.credentials())
	return err
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
	for key, cs := range h.sessions {
		errs = append(errs, cs.session.Close())
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

// ListTools implements drivers.MCPService.
func (h *handle) ListTools(ctx context.Context) ([]*drivers.MCPTool, error) {
	ctx, cancel := context.WithTimeout(ctx, h.config.timeout)
	defer cancel()

	var tools []*drivers.MCPTool
	err := h.withSession(ctx, h.credentials(), func(cs *cachedSession) error {
		h.mu.Lock()
		if cs.tools != nil && time.Now().Before(cs.toolsExpires) {
			tools = cs.tools
			h.mu.Unlock()
			return nil
		}
		h.mu.Unlock()

		res := []*drivers.MCPTool{}
		for t, err := range cs.session.Tools(ctx, nil) {
			if err != nil {
				return err
			}
			res = append(res, convertTool(t))
		}

		h.mu.Lock()
		cs.tools = res
		cs.toolsExpires = time.Now().Add(toolsCacheTTL)
		h.mu.Unlock()
		tools = res
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tools, nil
}

// CallTool implements drivers.MCPService.
func (h *handle) CallTool(ctx context.Context, name string, args map[string]any) (*drivers.MCPToolResult, error) {
	if args == nil {
		args = map[string]any{}
	}

	ctx, cancel := context.WithTimeout(ctx, h.config.timeout)
	defer cancel()

	var out *gomcp.CallToolResult
	err := h.withSession(ctx, h.credentials(), func(cs *cachedSession) error {
		var err error
		out, err = cs.session.CallTool(ctx, &gomcp.CallToolParams{Name: name, Arguments: args})
		return err
	})
	if err != nil {
		return nil, err
	}
	return h.convertResult(out), nil
}

// credentials returns the credentials for a request.
// All callers currently share the connector's headers, so they all use the same session.
func (h *handle) credentials() *credentials {
	return &credentials{headers: h.config.Headers}
}

// withSession runs fn with the session for the credentials, reconnecting once if the session was closed by the server.
func (h *handle) withSession(ctx context.Context, creds *credentials, fn func(cs *cachedSession) error) error {
	for attempt := range 2 {
		cs, err := h.connect(ctx, creds)
		if err != nil {
			return err
		}
		err = fn(cs)
		if errors.Is(err, gomcp.ErrConnectionClosed) && attempt == 0 {
			h.dropSession(creds.key, cs)
			continue
		}
		return err
	}
	return errors.New("failed to reconnect to MCP server")
}

// connect returns the session for the credentials, connecting if necessary.
func (h *handle) connect(ctx context.Context, creds *credentials) (*cachedSession, error) {
	h.mu.Lock()
	if cs, ok := h.sessions[creds.key]; ok {
		h.mu.Unlock()
		return cs, nil
	}
	h.mu.Unlock()

	client := gomcp.NewClient(&gomcp.Implementation{Name: clientName, Version: clientVersion}, nil)
	transport := &gomcp.StreamableClientTransport{
		Endpoint: h.config.URL,
		HTTPClient: &http.Client{
			Transport:     &headerRoundTripper{base: h.transport, headers: creds.headers},
			CheckRedirect: checkSameOriginRedirect,
		},
		MaxRetries:           1,
		DisableStandaloneSSE: true,
	}
	connectCtx, cancel := context.WithTimeout(ctx, h.config.connectTimeout)
	defer cancel()
	sess, err := client.Connect(connectCtx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MCP server: %w", err)
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if existing, ok := h.sessions[creds.key]; ok {
		// Another caller connected concurrently; keep theirs.
		_ = sess.Close()
		return existing, nil
	}
	cs := &cachedSession{session: sess}
	h.sessions[creds.key] = cs
	return cs, nil
}

// dropSession forgets the session if it is still the current one for the key, so the next call reconnects.
func (h *handle) dropSession(key string, cs *cachedSession) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sessions[key] == cs {
		_ = cs.session.Close()
		delete(h.sessions, key)
	}
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
