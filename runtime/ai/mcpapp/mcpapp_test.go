package mcpapp

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestMetaIsStructpbSafe asserts that the metadata attached to the resource and the tools only uses value types
// that structpb accepts, since tool _meta also flows into in-app LLM tool definitions through structpb.NewStruct.
func TestMetaIsStructpbSafe(t *testing.T) {
	a := newAsset([]byte("<html></html>"))

	_, err := structpb.NewStruct(WidgetResource(a).Meta)
	require.NoError(t, err)

	_, err = structpb.NewStruct(WidgetToolMeta(a))
	require.NoError(t, err)

	_, err = structpb.NewStruct(AppOnlyToolMeta([]string{VisibilityApp}))
	require.NoError(t, err)

	_, err = structpb.NewStruct((&ToolUI{Visibility: []string{VisibilityApp}}).Meta(nil))
	require.NoError(t, err)
}

func TestAssetURIContainsHash(t *testing.T) {
	a := newAsset([]byte("<html>v1</html>"))
	require.Len(t, a.Hash, 12)
	require.Equal(t, "ui://rill/widget/"+a.Hash+".html", a.URI)
	require.Contains(t, WidgetToolMeta(a)["ui"].(map[string]any)["resourceUri"], a.Hash)

	// A different bundle gets a different URI, so hosts that cache by URI pick up the new widget.
	b := newAsset([]byte("<html>v2</html>"))
	require.NotEqual(t, a.URI, b.URI)
}

func TestToolUIMeta(t *testing.T) {
	a := newAsset([]byte("<html></html>"))

	// A widget-rendering tool links to the resource, but only when the asset is available.
	widget := &ToolUI{}
	require.Equal(t, WidgetToolMeta(a), widget.Meta(a))
	require.Nil(t, widget.Meta(nil))

	// An app-only tool declares its visibility regardless of the asset.
	appOnly := &ToolUI{Visibility: []string{VisibilityApp}}
	require.Equal(t, map[string]any{"ui": map[string]any{"visibility": []any{"app"}}}, appOnly.Meta(nil))
	require.True(t, appOnly.AppOnly())
	require.False(t, widget.AppOnly())
	require.False(t, (&ToolUI{Visibility: []string{"model", VisibilityApp}}).AppOnly())

	var none *ToolUI
	require.Nil(t, none.Meta(a))
	require.False(t, none.AppOnly())
}

func TestToolWithMeta(t *testing.T) {
	spec := &mcp.Tool{Name: "create_chart", Meta: map[string]any{"openai/toolInvocation/invoking": "Creating chart..."}}

	// Empty meta returns the spec itself.
	require.Same(t, spec, ToolWithMeta(spec, nil))
	require.Same(t, spec, ToolWithMeta(spec, map[string]any{}))

	// Otherwise a shallow copy with the merged _meta, leaving the spec's own _meta untouched.
	a := newAsset([]byte("<html></html>"))
	got := ToolWithMeta(spec, WidgetToolMeta(a))
	require.NotSame(t, spec, got)
	require.Equal(t, "create_chart", got.Name)
	require.Equal(t, "Creating chart...", got.Meta["openai/toolInvocation/invoking"])
	require.Equal(t, a.URI, got.Meta["ui"].(map[string]any)["resourceUri"])
	require.NotContains(t, spec.Meta, "ui")

	// A spec without _meta gets one.
	got = ToolWithMeta(&mcp.Tool{Name: "chart_data"}, AppOnlyToolMeta([]string{VisibilityApp}))
	require.Equal(t, []any{"app"}, got.Meta["ui"].(map[string]any)["visibility"])
}

func TestWidgetDirOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(DirEnv, dir)

	// No file yet: no widget.
	_, ok := Widget()
	require.False(t, ok)

	// The file is picked up without a restart.
	path := filepath.Join(dir, "widget.html")
	require.NoError(t, os.WriteFile(path, []byte("<html>v1</html>"), 0o644))
	a, ok := Widget()
	require.True(t, ok)
	require.Equal(t, "<html>v1</html>", a.HTML)

	// A change to the file (with a new mtime) is re-read and gets a new URI.
	require.NoError(t, os.WriteFile(path, []byte("<html>v2</html>"), 0o644))
	require.NoError(t, os.Chtimes(path, time.Now(), time.Now().Add(time.Second)))
	b, ok := Widget()
	require.True(t, ok)
	require.Equal(t, "<html>v2</html>", b.HTML)
	require.NotEqual(t, a.URI, b.URI)
}

func TestRegisterAndCapabilities(t *testing.T) {
	a := newAsset([]byte("<html></html>"))
	srv := mcp.NewServer(&mcp.Implementation{Name: "test"}, &mcp.ServerOptions{Capabilities: ServerCapabilities()})
	Register(srv, a)

	client := mcp.NewClient(&mcp.Implementation{Name: "client"}, nil)
	ct, st := mcp.NewInMemoryTransports()
	_, err := srv.Connect(t.Context(), st, nil)
	require.NoError(t, err)
	cs, err := client.Connect(t.Context(), ct, nil)
	require.NoError(t, err)
	defer cs.Close()

	caps := cs.InitializeResult().Capabilities
	require.NotNil(t, caps.Logging)
	require.Contains(t, caps.Extensions, UIExtension)
	require.NotNil(t, caps.Resources)

	list, err := cs.ListResources(t.Context(), &mcp.ListResourcesParams{})
	require.NoError(t, err)
	require.Len(t, list.Resources, 1)
	require.Equal(t, a.URI, list.Resources[0].URI)
	require.Equal(t, MIMEType, list.Resources[0].MIMEType)
	require.Equal(t, true, list.Resources[0].Meta["ui"].(map[string]any)["prefersBorder"])

	res, err := cs.ReadResource(t.Context(), &mcp.ReadResourceParams{URI: a.URI})
	require.NoError(t, err)
	require.Len(t, res.Contents, 1)
	require.Equal(t, MIMEType, res.Contents[0].MIMEType)
	require.Equal(t, a.HTML, res.Contents[0].Text)
}

// TestRegisterServesStaleURIs asserts that the widget is served under its current URI and under the URI of any
// earlier build, since hosts cache the tool list and keep requesting the URI they last saw after a release.
func TestRegisterServesStaleURIs(t *testing.T) {
	ctx := t.Context()
	a := newAsset([]byte("<html>current</html>"))

	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	Register(srv, a)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := srv.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// The current URI is listed and readable.
	list, err := clientSession.ListResources(ctx, nil)
	require.NoError(t, err)
	require.Len(t, list.Resources, 1)
	require.Equal(t, a.URI, list.Resources[0].URI)
	require.Equal(t, MIMEType, list.Resources[0].MIMEType)

	res, err := clientSession.ReadResource(ctx, &mcp.ReadResourceParams{URI: a.URI})
	require.NoError(t, err)
	require.Len(t, res.Contents, 1)
	require.Equal(t, a.URI, res.Contents[0].URI)
	require.Equal(t, MIMEType, res.Contents[0].MIMEType)
	require.Equal(t, a.HTML, res.Contents[0].Text)

	// A stale URI from an earlier build gets the current bundle, echoing the requested URI.
	stale := "ui://rill/widget/000000000000.html"
	res, err = clientSession.ReadResource(ctx, &mcp.ReadResourceParams{URI: stale})
	require.NoError(t, err)
	require.Len(t, res.Contents, 1)
	require.Equal(t, stale, res.Contents[0].URI)
	require.Equal(t, MIMEType, res.Contents[0].MIMEType)
	require.Equal(t, a.HTML, res.Contents[0].Text)
	require.NotNil(t, res.Contents[0].Meta["ui"])

	// Anything else is not found.
	_, err = clientSession.ReadResource(ctx, &mcp.ReadResourceParams{URI: "ui://rill/other.html"})
	require.Error(t, err)
}
