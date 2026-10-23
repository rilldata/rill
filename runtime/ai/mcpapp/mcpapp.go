// Package mcpapp serves the MCP Apps widget that renders Rill charts inside MCP hosts such as Claude and ChatGPT.
//
// MCP Apps (the io.modelcontextprotocol/ui extension) lets a server register a resource with a ui:// URI whose body is a
// self-contained HTML document. A tool links to it through _meta.ui.resourceUri, and the host renders the HTML in a sandboxed
// iframe, pushing the tool call's result into it. This package owns the embedded widget bundle, the resource and tool
// metadata that describe it, and the server capability that advertises support for the extension.
//
// The package has no dependency on the runtime so that tool specs can be built without one.
package mcpapp

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	// MIMEType is the MIME type of MCP Apps HTML resources.
	MIMEType = "text/html;profile=mcp-app"
	// ResourceName is the name of the widget resource.
	ResourceName = "rill-widget"
	// UIExtension is the identifier of the MCP Apps extension in capability negotiation.
	UIExtension = "io.modelcontextprotocol/ui"
	// VisibilityApp marks a tool as callable only by the widget (through the host), not by the model.
	VisibilityApp = "app"

	// DirEnv is an environment variable naming a directory to load the widget from instead of the embedded copy.
	// The file is re-read whenever it changes, which gives a hot reload loop during widget development.
	DirEnv = "RILL_MCP_APP_DIR"

	widgetFile = "widget.html"
	uriPrefix  = "ui://rill/widget/"
	// uriTemplate matches every widget URI, including ones from earlier builds (see Register).
	uriTemplate = uriPrefix + "{build}.html"
)

//go:embed all:embed
var embedFS embed.FS

// Asset is a loaded widget bundle.
type Asset struct {
	// Name is the file name of the bundle.
	Name string
	// Hash is the first 12 hex characters of the SHA-256 of HTML.
	// The URI embeds it so that hosts, which cache resources by URI, refetch the widget exactly when the bundle changes.
	Hash string
	// URI is the ui:// URI the widget resource is served at.
	URI string
	// HTML is the self-contained widget document.
	HTML string
}

// ToolUI describes how a tool takes part in the MCP Apps widget.
type ToolUI struct {
	// Visibility lists who may call the tool. Empty means the model: the tool renders the widget with its result.
	// VisibilityApp means only the widget may call it (through the host), for example to refetch data.
	Visibility []string
}

// AppOnly reports whether only the widget may call the tool: its visibility is exactly VisibilityApp.
// Calls to such tools are plumbing (the widget refetching data) rather than conversation turns.
func (u *ToolUI) AppOnly() bool {
	return u != nil && slices.Equal(u.Visibility, []string{VisibilityApp})
}

// Meta returns the _meta.ui entry for a tool with this UI, or nil when the tool has nothing to advertise.
// Widget-rendering tools link to the widget resource, which requires the asset; app-only tools only declare their visibility.
func (u *ToolUI) Meta(a *Asset) map[string]any {
	if u == nil {
		return nil
	}
	if len(u.Visibility) > 0 {
		return AppOnlyToolMeta(u.Visibility)
	}
	if a == nil {
		return nil
	}
	return WidgetToolMeta(a)
}

// widgetCache memoizes the loaded widget. The embedded copy never changes; a DirEnv override is re-read when its mtime changes.
var widgetCache struct {
	sync.Mutex
	embedded     *Asset
	embeddedDone bool
	overridePath string
	overrideTime time.Time
	override     *Asset
}

// Widget returns the widget bundle, or false when none is available.
// The bundle comes from DirEnv when set, otherwise from the copy embedded at build time (see embed/README.md).
func Widget() (*Asset, bool) {
	widgetCache.Lock()
	defer widgetCache.Unlock()

	if dir := os.Getenv(DirEnv); dir != "" {
		path := filepath.Join(dir, widgetFile)
		info, err := os.Stat(path)
		if err != nil {
			return nil, false
		}
		if widgetCache.override == nil || widgetCache.overridePath != path || !widgetCache.overrideTime.Equal(info.ModTime()) {
			html, err := os.ReadFile(path)
			if err != nil {
				return nil, false
			}
			widgetCache.override = newAsset(html)
			widgetCache.overridePath = path
			widgetCache.overrideTime = info.ModTime()
		}
		return widgetCache.override, true
	}

	if !widgetCache.embeddedDone {
		widgetCache.embeddedDone = true
		html, err := embedFS.ReadFile("embed/" + widgetFile)
		if err == nil {
			widgetCache.embedded = newAsset(html)
		}
	}
	return widgetCache.embedded, widgetCache.embedded != nil
}

func newAsset(html []byte) *Asset {
	sum := sha256.Sum256(html)
	hash := hex.EncodeToString(sum[:])[:12]
	return &Asset{
		Name: widgetFile,
		Hash: hash,
		URI:  uriPrefix + hash + ".html",
		HTML: string(html),
	}
}

// WidgetResource returns the MCP resource that serves the widget.
// Its _meta follows the MCP Apps spec plus the ChatGPT extras: a CSP that only allows Claude's font assets,
// a host-drawn border, inline and fullscreen display modes, and a description that keeps the model from restating the chart.
// Nested values are only map[string]any and []any so that they can also be converted with structpb.
func WidgetResource(a *Asset) *mcp.Resource {
	return &mcp.Resource{
		URI:         a.URI,
		Name:        ResourceName,
		Title:       resourceTitle,
		Description: resourceDescription,
		MIMEType:    MIMEType,
		Size:        int64(len(a.HTML)),
		Meta:        resourceMeta(),
	}
}

const (
	resourceTitle       = "Rill chart"
	resourceDescription = "Interactive Rill chart widget rendered from the result of create_chart."
)

// resourceMeta returns the _meta of the widget resource; a fresh map on every call since the SDK may retain it.
func resourceMeta() map[string]any {
	return map[string]any{
		"ui": map[string]any{
			"csp": map[string]any{
				"connectDomains":  []any{},
				"resourceDomains": []any{"https://assets.claude.ai"},
				"frameDomains":    []any{},
				"baseUriDomains":  []any{},
			},
			"prefersBorder": true,
		},
		"openai/ui": map[string]any{
			"availableDisplayModes": []any{"inline", "fullscreen"},
		},
		"openai/widgetDescription": "Interactive Rill chart. Summarize the insight in one or two sentences; do not restate the numbers shown in the chart.",
	}
}

// WidgetToolMeta returns the _meta entry that links a tool to the widget resource.
func WidgetToolMeta(a *Asset) map[string]any {
	return map[string]any{
		"ui": map[string]any{
			"resourceUri": a.URI,
		},
	}
}

// AppOnlyToolMeta returns the _meta entry that restricts a tool to the given callers (see VisibilityApp).
func AppOnlyToolMeta(visibility []string) map[string]any {
	vis := make([]any, len(visibility))
	for i, v := range visibility {
		vis[i] = v
	}
	return map[string]any{
		"ui": map[string]any{
			"visibility": vis,
		},
	}
}

// ToolWithMeta returns a shallow copy of a tool spec with meta merged into its _meta, or the spec itself when meta is empty.
// The copy gets its own _meta map, so the spec's _meta (which also flows into the in-app tool definitions) is left untouched.
func ToolWithMeta(spec *mcp.Tool, meta map[string]any) *mcp.Tool {
	if len(meta) == 0 {
		return spec
	}
	res := *spec
	res.Meta = make(map[string]any, len(spec.Meta)+len(meta))
	maps.Copy(res.Meta, spec.Meta)
	maps.Copy(res.Meta, meta)
	return &res
}

// Register adds the widget resource to the server.
// The current URI is listed as a resource. A template additionally serves the current bundle under any
// widget URI, because hosts cache the tool list and keep requesting the URI of the build they last saw;
// after a release that would otherwise fail until they refresh. The response echoes the requested URI.
func Register(srv *mcp.Server, a *Asset) {
	read := func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		if !strings.HasPrefix(req.Params.URI, uriPrefix) || !strings.HasSuffix(req.Params.URI, ".html") {
			return nil, fmt.Errorf("unknown resource %q", req.Params.URI)
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      req.Params.URI,
				MIMEType: MIMEType,
				Text:     a.HTML,
				Meta:     resourceMeta(),
			}},
		}, nil
	}
	srv.AddResource(WidgetResource(a), read)
	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: uriTemplate,
		Name:        ResourceName,
		Title:       resourceTitle,
		Description: resourceDescription,
		MIMEType:    MIMEType,
		Meta:        resourceMeta(),
	}, read)
}

// ServerCapabilities returns the server capabilities to use when the widget is served:
// the SDK's default logging capability plus the MCP Apps extension with the MIME types the server can render.
// Tools and resources capabilities are still inferred by the SDK from what is registered.
func ServerCapabilities() *mcp.ServerCapabilities {
	caps := &mcp.ServerCapabilities{
		Logging: &mcp.LoggingCapabilities{},
	}
	caps.AddExtension(UIExtension, map[string]any{
		"mimeTypes": []any{MIMEType},
	})
	return caps
}
