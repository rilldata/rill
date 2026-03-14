package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai/instructions"
)

const DevelopFileName = "develop_file"

type DevelopFile struct {
	Runtime *runtime.Runtime
}

var _ Tool[*DevelopFileArgs, *DevelopFileResult] = (*DevelopFile)(nil)

type DevelopFileArgs struct {
	Path   string `json:"path" jsonschema:"The path of a .yaml or .sql file to create, update or delete."`
	Type   string `json:"type,omitempty" jsonschema:"Type of Rill file to develop (optional, but recommended if known). Options: rill.yaml, .env, connector, model, metrics_view, explore, canvas, theme, api, alert, report."`
	Prompt string `json:"prompt" jsonschema:"A detailed description of how to develop the file. Include any relevant details assuming no prior context except the path's current content and status (if any)."`
}

type DevelopFileResult struct {
	Response string `json:"response" jsonschema:"The developer's response summarizing the actions taken."`
}

func (t *DevelopFile) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        DevelopFileName,
		Title:       "Develop file",
		Description: "Developer agent that creates, edits or deletes a single Rill project file based on a prompt. It has no prior context from the conversation, but has deep knowledge of Rill project development and best practices.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Developing file...",
			"openai/toolInvocation/invoked":  "Developed file",
		},
	}
}

func (t *DevelopFile) CheckAccess(ctx context.Context) (bool, error) {
	return checkDeveloperAccess(ctx, t.Runtime, true)
}

func (t *DevelopFile) Handler(ctx context.Context, args *DevelopFileArgs) (*DevelopFileResult, error) {
	// Validate input
	if args.Path == "" || args.Prompt == "" {
		return nil, fmt.Errorf("invalid input: path and prompt are required")
	}
	if !strings.HasPrefix(args.Path, "/") {
		args.Path = "/" + args.Path
	}

	// Prepare the system prompts
	generalInstructions, err := instructions.Load("development.md", instructions.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to load developer agent system prompt: %w", err)
	}
	var resourceInstructions *instructions.Instruction
	switch args.Type {
	case "", ".env", "api", "alert", "report":
		// These types currently don't have additional resource-specific instructions
	case "rill.yaml":
		resourceInstructions, err = instructions.Load("resources/rillyaml.md", instructions.Options{})
		if err != nil {
			return nil, fmt.Errorf("failed to load developer agent resource-specific system prompt: %w", err)
		}
	case "connector", "model", "metrics_view", "explore", "canvas", "theme":
		resourceInstructions, err = instructions.Load(fmt.Sprintf("resources/%s.md", args.Type), instructions.Options{})
		if err != nil {
			return nil, fmt.Errorf("failed to load developer agent resource-specific system prompt: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid input: unsupported resource type %q", args.Type)
	}

	// Prepare the user prompt
	userPrompt, err := t.userPrompt(ctx, args)
	if err != nil {
		return nil, err
	}

	// Pre-invoke some tool calls
	s := GetSession(ctx)
	_, err = s.CallTool(ctx, RoleAssistant, ListFilesName, nil, &ListFilesArgs{})
	if err != nil {
		return nil, err
	}
	_, err = s.CallTool(ctx, RoleAssistant, ProjectStatusName, nil, &ProjectStatusArgs{})
	if err != nil {
		return nil, err
	}
	_, _ = s.CallTool(ctx, RoleAssistant, ReadFileName, nil, &ReadFileArgs{
		Path: args.Path,
	})
	if ctx.Err() != nil { // Ignore tool error since the file may not exist
		return nil, ctx.Err()
	}

	// Build initial completion messages
	messages := []*aiv1.CompletionMessage{NewTextCompletionMessage(RoleSystem, generalInstructions.Body)}
	if resourceInstructions != nil {
		messages = append(messages, NewTextCompletionMessage(RoleSystem, resourceInstructions.Body))
	}
	messages = append(messages, NewTextCompletionMessage(RoleUser, userPrompt))
	messages = append(messages, s.NewCompletionMessages(s.MessagesWithResults(FilterByParent(s.ID())))...)

	// Run an LLM tool call loop
	var response string
	err = s.Complete(ctx, "File developer loop", &response, &CompleteOptions{
		Messages: messages,
		Tools: []string{
			SearchFilesName,
			ReadFileName,
			WriteFileName,
			ListBucketsName,
			ListBucketObjectsName,
			ListTablesName,
			ShowTableName,
			QuerySQLName,
		},
		MaxIterations: 10,
		UnwrapCall:    true,
	})
	if err != nil {
		return nil, err
	}

	return &DevelopFileResult{
		Response: response,
	}, nil
}

func (t *DevelopFile) userPrompt(ctx context.Context, args *DevelopFileArgs) (string, error) {
	// Get default OLAP info
	olapInfo, err := defaultOLAPInfo(ctx, t.Runtime, GetSession(ctx).InstanceID())
	if err != nil {
		return "", err
	}

	// Prepare template data.
	session := GetSession(ctx)
	data := map[string]any{
		"path":              args.Path,
		"type":              args.Type,
		"prompt":            args.Prompt,
		"ai_instructions":   session.ProjectInstructions(),
		"default_olap_info": olapInfo,
	}

	// Generate the user prompt
	return executeTemplate(`
You should develop a Rill project file based on the following task description:
- Develop file at path: {{ .path }}
{{ if .type }}- The file should be of type: {{ .type }}{{ end }}
- Task description: {{ .prompt }}

Here is some important context:
- You are running as a sub-agent of a larger developer agent. Stay aligned on your specific task and avoid extra discovery.
- When you call 'write_file', if it returns a parse or reconcile error, do your best to fix the issue and try again. If you think the error is unrelated to the current path, let the parent agent know to handle it.

Here is some additional context that may or may not be relevant to your task:
- Info about the project's default OLAP connector: {{ .default_olap_info }}.
{{ if .ai_instructions }}- The user has configured global additional instructions for you. They may not relate to the current request, and may not even relate to your work as a data engineer agent. Only use them if you find them relevant. They are: {{ .ai_instructions }}{{ end }}
`, data)
}
