package ai

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
	"gopkg.in/yaml.v3"
)

// CreatePersonalCanvasName is the MCP tool identifier surfaced to the model.
const CreatePersonalCanvasName = "create_personal_canvas"

// CreatePersonalCanvas is an AI tool that creates a personal canvas dashboard for the current user.
// The canvas is stored under "personal/canvases/{user_id}/{slug}.yaml" with owner-only access.
//
// In Rill Developer the runtime writes directly to the project repo. In Rill Cloud the runtime is
// not directly editable by viewers; for cloud, callers use the dedicated admin RPC
// (CreatePersonalVirtualFile) and the in-app dialog rather than this tool.
type CreatePersonalCanvas struct {
	Runtime *runtime.Runtime
}

var _ Tool[*CreatePersonalCanvasArgs, *CreatePersonalCanvasResult] = (*CreatePersonalCanvas)(nil)

type CreatePersonalCanvasArgs struct {
	DisplayName string `json:"display_name" jsonschema:"Human-readable display name for the canvas."`
	YAML        string `json:"yaml,omitempty" jsonschema:"Optional canvas YAML body. If omitted, a blank canvas is created. The 'type: canvas' field is required."`
}

type CreatePersonalCanvasResult struct {
	Name        string `json:"name" jsonschema:"The generated canvas resource name."`
	DisplayName string `json:"display_name" jsonschema:"The display name of the created canvas."`
	Path        string `json:"path" jsonschema:"The virtual file path the canvas was written to."`
}

func (t *CreatePersonalCanvas) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        CreatePersonalCanvasName,
		Title:       "Create personal canvas",
		Description: "Creates a personal canvas dashboard that is only visible to the calling user. The canvas is stored as a virtual file scoped to the user's identity. Use this when the user asks the assistant to build or generate a personal/private dashboard for themselves.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  false,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    false,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Creating personal canvas...",
			"openai/toolInvocation/invoked":  "Created personal canvas",
		},
	}
}

func (t *CreatePersonalCanvas) CheckAccess(ctx context.Context) (bool, error) {
	s := GetSession(ctx)
	if !s.Claims().Can(runtime.UseAI) {
		return false, nil
	}

	// The tool writes a YAML file to the project, which requires EditRepo today.
	// In Rill Developer the local user has EditRepo; in Rill Cloud viewers do not, and cloud
	// creation routes through the admin RPC + UI instead.
	if !s.Claims().Can(runtime.EditRepo) {
		return false, nil
	}

	// Must have a user identity to attribute ownership.
	if userID, _ := s.Claims().UserAttributes["id"].(string); userID == "" && !s.Claims().SkipChecks {
		return false, nil
	}

	ff, err := t.Runtime.FeatureFlags(ctx, s.InstanceID(), s.Claims())
	if err != nil {
		return false, err
	}
	return ff["personal_canvases"], nil
}

func (t *CreatePersonalCanvas) Handler(ctx context.Context, args *CreatePersonalCanvasArgs) (*CreatePersonalCanvasResult, error) {
	s := GetSession(ctx)

	displayName := strings.TrimSpace(args.DisplayName)
	if displayName == "" {
		return nil, fmt.Errorf("display_name is required")
	}

	ownerID, _ := s.Claims().UserAttributes["id"].(string)
	if ownerID == "" {
		ownerID = "local" // Rill Developer fallback
	}

	var body []byte
	var err error
	if args.YAML != "" {
		body, err = sanitizePersonalCanvasYAML([]byte(args.YAML), displayName, ownerID)
		if err != nil {
			return nil, fmt.Errorf("invalid YAML: %w", err)
		}
	} else {
		body, err = blankPersonalCanvasYAML(displayName, ownerID)
		if err != nil {
			return nil, fmt.Errorf("failed to build blank canvas: %w", err)
		}
	}

	name := randomPersonalCanvasName(displayName)
	filePath := personalCanvasPath(ownerID, name)

	err = t.Runtime.PutFile(ctx, s.InstanceID(), filePath, strings.NewReader(string(body)), true, true)
	if err != nil {
		return nil, fmt.Errorf("failed to write personal canvas: %w", err)
	}

	return &CreatePersonalCanvasResult{
		Name:        name,
		DisplayName: displayName,
		Path:        filePath,
	}, nil
}

// sanitizePersonalCanvasYAML decodes the agent-provided YAML, enforces type=canvas, sets the
// display name, and injects owner-only annotations.
func sanitizePersonalCanvasYAML(data []byte, displayName, ownerID string) ([]byte, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if doc == nil {
		doc = map[string]any{}
	}

	docType, _ := doc["type"].(string)
	if docType != "" && docType != "canvas" {
		return nil, fmt.Errorf("type must be canvas, got %q", docType)
	}
	doc["type"] = "canvas"

	if displayName != "" {
		doc["display_name"] = displayName
	}

	annotations, _ := doc["annotations"].(map[string]any)
	if annotations == nil {
		annotations = map[string]any{}
	}
	annotations["admin_owner_user_id"] = ownerID
	annotations["admin_managed"] = true
	annotations["admin_nonce"] = time.Now().Format(time.RFC3339Nano)
	doc["annotations"] = annotations

	return yaml.Marshal(doc)
}

// blankPersonalCanvasYAML returns a minimal canvas YAML for a new personal canvas.
func blankPersonalCanvasYAML(displayName, ownerID string) ([]byte, error) {
	doc := map[string]any{
		"type":         "canvas",
		"display_name": displayName,
		"annotations": map[string]any{
			"admin_owner_user_id": ownerID,
			"admin_managed":       true,
			"admin_nonce":         time.Now().Format(time.RFC3339Nano),
		},
		"rows": []any{},
	}
	return yaml.Marshal(doc)
}

func personalCanvasPath(ownerID, name string) string {
	return "/" + path.Join("personal", "canvases", ownerID, name+".yaml")
}

var personalCanvasDashCharsRegexp = regexp.MustCompile(`[ _]+`)

var personalCanvasExcludeCharsRegexp = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

func randomPersonalCanvasName(displayName string) string {
	name := personalCanvasDashCharsRegexp.ReplaceAllString(displayName, "-")
	name = personalCanvasExcludeCharsRegexp.ReplaceAllString(name, "")
	name = strings.ToLower(name)
	name = strings.Trim(name, "-")
	if name == "" {
		return uuid.New().String()
	}
	return name + "-" + uuid.New().String()[0:8]
}
