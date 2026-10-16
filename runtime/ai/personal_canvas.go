package ai

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai/instructions"
	"gopkg.in/yaml.v3"
)

const DevelopPersonalCanvasName = "develop_personal_canvas"

// DevelopPersonalCanvas creates or edits a personal canvas dashboard of the current user, delegating the authoring to a sub-agent.
// Unlike develop_file, it doesn't need edit access to the project: personal canvases are owned by the user and stored by the admin service, outside the project's repo.
type DevelopPersonalCanvas struct {
	Runtime *runtime.Runtime
}

var _ Tool[*DevelopPersonalCanvasArgs, *DevelopPersonalCanvasResult] = (*DevelopPersonalCanvas)(nil)

type DevelopPersonalCanvasArgs struct {
	Prompt string `json:"prompt" jsonschema:"A detailed description of the dashboard to build or the changes to make. Include the metrics, breakdowns, time ranges and filters the user asked for, assuming no prior context. To start from an existing project dashboard, name it here."`
	Name   string `json:"name,omitempty" jsonschema:"Optional name of one of the user's existing personal canvas dashboards to edit, as listed by list_canvases with personal: true. Leave empty to create a new one."`
}

type DevelopPersonalCanvasResult struct {
	Response string `json:"response" jsonschema:"The developer's response summarizing the dashboard, including its name and URL."`
}

func (t *DevelopPersonalCanvas) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        DevelopPersonalCanvasName,
		Title:       "Develop personal dashboard",
		Description: "Creates a canvas dashboard for the current user, or edits one of the user's own, and saves it as a personal dashboard that only they can see until they share it from the UI. Use it when the user asks to create, build, save or change a dashboard for themselves. It can't share or delete dashboards, nor change the project's dashboards. To edit a dashboard, find its name with list_canvases; if the user asks to change a dashboard you can't find, ask them which one instead of creating a new one. Give the user the dashboard's URL from the response.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  false,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    false,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Developing dashboard...",
			"openai/toolInvocation/invoked":  "Developed dashboard",
		},
	}
}

func (t *DevelopPersonalCanvas) CheckAccess(ctx context.Context) (bool, error) {
	return checkPersonalCanvasAccess(ctx, t.Runtime)
}

func (t *DevelopPersonalCanvas) Handler(ctx context.Context, args *DevelopPersonalCanvasArgs) (*DevelopPersonalCanvasResult, error) {
	if args.Prompt == "" {
		return nil, errors.New("invalid input: prompt is required")
	}
	s := GetSession(ctx)

	// Load the current YAML of the canvas to edit, after checking the user owns it.
	var current string
	if args.Name != "" {
		var err error
		current, err = readPersonalCanvas(ctx, t.Runtime, args.Name)
		if err != nil {
			return nil, err
		}
	}

	// Prepare the system prompts
	canvasInstructions, err := instructions.Load("resources/canvas.md", instructions.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to load canvas instructions: %w", err)
	}

	// Prepare the user prompt. The field names come from the same helper as develop_file, so both see the metrics views the same way.
	metricsViewsInfo, err := (&DevelopFile{Runtime: t.Runtime}).metricsViewsInfo(ctx)
	if err != nil {
		return nil, err
	}
	// get_canvas is not available to every client (e.g. MCP clients), so only mention it when the sub-agent can call it.
	canGetCanvas, err := (&GetCanvas{Runtime: t.Runtime}).CheckAccess(ctx)
	if err != nil {
		return nil, err
	}
	tools := []string{WritePersonalCanvasName, GetMetricsViewName}
	if canGetCanvas {
		tools = append(tools, ListCanvasesName, GetCanvasName)
	}

	userPrompt, err := executeTemplate(`
You should develop a personal canvas dashboard based on the following task description:
{{ if .name }}- Edit the user's existing personal canvas named: {{ .name }}{{ else }}- Create a new personal canvas{{ end }}
- Task description: {{ .prompt }}

Here is some important context:
- You are running as a sub-agent of an analyst agent that is talking to a business user.
- Save the dashboard with 'write_personal_canvas'. It is stored outside the project's files, so there is no file path. {{ if .name }}Pass name "{{ .name }}" to edit it.{{ else }}Leave name empty the first time, then pass the name it returns on every later call so you edit the same dashboard instead of creating another one.{{ end }}
- Don't set annotations: they are ignored. The dashboard is owned by the user and private to them.
- Only use the exact measure and dimension names of the metrics views listed below. If a metrics view is not listed, call 'get_metrics_view' BEFORE writing. Never guess field names based on display names or the task description.
{{ if .get_canvas }}- To start from an existing project dashboard, find its name with 'list_canvases', call 'get_canvas' and reproduce its components in the new dashboard. The project dashboard itself is never changed.
{{ end }}- When 'write_personal_canvas' returns a parse or reconcile error, fix the YAML and save it again.
- Reply with the dashboard's display name and its URL.
{{ if .current }}
The current YAML of the dashboard is:
~~~yaml
{{ .current }}
~~~
{{ end }}
{{ if .metrics_views_info }}The project's metrics views and their exact field names:
{{ .metrics_views_info }}{{ end }}
{{ if .ai_instructions }}The user has configured global additional instructions. Only use them if relevant to building the dashboard: {{ .ai_instructions }}{{ end }}
`, map[string]any{
		"name":               args.Name,
		"prompt":             args.Prompt,
		"current":            current,
		"metrics_views_info": metricsViewsInfo,
		"ai_instructions":    s.ProjectInstructions(),
		"get_canvas":         canGetCanvas,
	})
	if err != nil {
		return nil, err
	}

	messages := []*aiv1.CompletionMessage{
		NewTextCompletionMessage(RoleSystem, canvasInstructions.Body),
		NewTextCompletionMessage(RoleUser, userPrompt),
	}
	messages = append(messages, s.NewCompletionMessages(s.MessagesWithResults(FilterByParent(s.ParentID)))...)

	var response string
	err = s.Complete(ctx, "Personal canvas developer loop", &response, &CompleteOptions{
		Messages:      messages,
		Tools:         tools,
		MaxIterations: 10,
		UnwrapCall:    true,
	})
	if err != nil {
		return nil, err
	}

	return &DevelopPersonalCanvasResult{
		Response: response,
	}, nil
}

const WritePersonalCanvasName = "write_personal_canvas"

// WritePersonalCanvas saves a personal canvas of the current user through the admin service, and returns its parse and reconcile status.
// It is only available to the develop_personal_canvas sub-agent.
type WritePersonalCanvas struct {
	Runtime *runtime.Runtime
}

var _ Tool[*WritePersonalCanvasArgs, *WritePersonalCanvasResult] = (*WritePersonalCanvas)(nil)

type WritePersonalCanvasArgs struct {
	Name        string `json:"name,omitempty" jsonschema:"Name of the personal canvas to edit. Leave empty to create a new one."`
	DisplayName string `json:"display_name,omitempty" jsonschema:"Display name of the dashboard. Required when creating a new one."`
	YAML        string `json:"yaml" jsonschema:"The full YAML of the canvas dashboard, with type: canvas."`
}

type WritePersonalCanvasResult struct {
	Name          string           `json:"name" jsonschema:"Name of the saved personal canvas. Pass it to edit the same dashboard."`
	URL           string           `json:"url" jsonschema:"URL where the user can open the dashboard."`
	Resources     []map[string]any `json:"resources,omitempty" jsonschema:"Status of the dashboard resource, including any reconcile error."`
	ParseError    string           `json:"parse_error,omitempty" jsonschema:"Parse error encountered when parsing the YAML, if any."`
	ParseWarnings []string         `json:"parse_warnings,omitempty" jsonschema:"Parse warnings encountered when parsing the YAML, if any."`
}

func (t *WritePersonalCanvas) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        WritePersonalCanvasName,
		Title:       "Write personal dashboard",
		Description: "Saves a canvas dashboard as a personal dashboard of the current user, creating it or overwriting one of the user's own. It waits for the dashboard to reconcile and returns its name, URL and any errors encountered.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(true),
			IdempotentHint:  true,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    false,
		},
	}
}

func (t *WritePersonalCanvas) CheckAccess(ctx context.Context) (bool, error) {
	// Only expose it inside develop_personal_canvas, so MCP clients and the analyst don't write canvases without the canvas instructions.
	if GetSession(ctx).ancestorCall(DevelopPersonalCanvasName) == nil {
		return false, nil
	}
	return checkPersonalCanvasAccess(ctx, t.Runtime)
}

func (t *WritePersonalCanvas) Handler(ctx context.Context, args *WritePersonalCanvasArgs) (*WritePersonalCanvasResult, error) {
	if args.YAML == "" {
		return nil, errors.New("invalid input: yaml is required")
	}
	if args.Name == "" && args.DisplayName == "" {
		return nil, errors.New("invalid input: display_name is required when creating a dashboard")
	}
	s := GetSession(ctx)

	admin, release, err := t.Runtime.Admin(ctx, s.InstanceID())
	if err != nil {
		return nil, err
	}
	defer release()

	c, err := admin.SavePersonalCanvas(ctx, s.Claims().UserID, args.Name, args.DisplayName, args.YAML)
	if err != nil {
		return nil, fmt.Errorf("failed to save personal canvas: %w", err)
	}

	resources, parseErr, parseWarnings, err := (&WriteFile{Runtime: t.Runtime}).reconcileAndGetStatus(ctx, personalCanvasPath(c.Name), true)
	if err != nil {
		return nil, err
	}

	return &WritePersonalCanvasResult{
		Name:          c.Name,
		URL:           c.URL,
		Resources:     resources,
		ParseError:    parseErr,
		ParseWarnings: parseWarnings,
	}, nil
}

// checkPersonalCanvasAccess checks that the session belongs to a user who may use AI in a prod deployment with personal canvases enabled.
// Personal canvases are stored in the prod environment, so other deployments (like editable branch deployments) can't save them.
// Users are told apart from services by their email attribute, which only user JWTs carry.
// The admin service additionally rejects owners that aren't users with access to the project, such as embed users with custom attributes.
func checkPersonalCanvasAccess(ctx context.Context, rt *runtime.Runtime) (bool, error) {
	s := GetSession(ctx)
	claims := s.Claims()
	if !claims.Can(runtime.UseAI) || claims.UserID == "" {
		return false, nil
	}
	if email, _ := claims.UserAttributes["email"].(string); email == "" {
		return false, nil
	}

	inst, err := rt.Instance(ctx, s.InstanceID())
	if err != nil {
		return false, err
	}
	if inst.Environment != "prod" {
		return false, nil
	}

	ff, err := rt.FeatureFlags(ctx, s.InstanceID(), claims)
	if err != nil {
		return false, err
	}
	return ff["personal_canvases"], nil
}

// readPersonalCanvas returns the YAML of a personal canvas owned by the session's user.
func readPersonalCanvas(ctx context.Context, rt *runtime.Runtime, name string) (string, error) {
	s := GetSession(ctx)
	repo, release, err := rt.Repo(ctx, s.InstanceID())
	if err != nil {
		return "", err
	}
	defer release()

	data, err := repo.Get(ctx, personalCanvasPath(name))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("personal canvas %q not found: call list_canvases for the names of the user's personal canvases", name)
		}
		return "", err
	}

	var doc struct {
		Annotations map[string]any `yaml:"annotations"`
	}
	if err := yaml.Unmarshal([]byte(data), &doc); err != nil {
		return "", fmt.Errorf("failed to parse personal canvas %q: %w", name, err)
	}
	if doc.Annotations["admin_owner_user_id"] != s.Claims().UserID {
		return "", fmt.Errorf("personal canvas %q not found: call list_canvases for the names of the user's personal canvases", name)
	}
	return data, nil
}

// personalCanvasPath returns the path where the runtime's repo presents a personal canvas stored by the admin service.
func personalCanvasPath(name string) string {
	return path.Join("/__virtual__", "personal", name+".yaml")
}
