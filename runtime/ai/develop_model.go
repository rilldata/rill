package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

type DevelopModel struct {
	Runtime *runtime.Runtime
}

var _ Tool[*DevelopModelArgs, *DevelopModelResult] = (*DevelopModel)(nil)

type DevelopModelArgs struct {
	Path   string `json:"path" jsonschema:"The path of a .yaml file in which to create or update a Rill model definition."`
	Prompt string `json:"prompt" jsonschema:"A description of what the model should do, i.e. what kind of data it should ingest and how it transform and output it."`
}

type DevelopModelResult struct {
	ModelName string `json:"model_name" jsonschema:"The name of the developed Rill model."`
}

func (t *DevelopModel) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        "develop_model",
		Title:       "Develop model",
		Description: "Agent that develops a single Rill model.",
	}
}

func (t *DevelopModel) CheckAccess(ctx context.Context) bool {
	// NOTE: Disabled pending further improvements
	// s := GetSession(ctx)
	// return s.Claims().Can(runtime.EditRepo)
	return false
}

func (t *DevelopModel) Handler(ctx context.Context, args *DevelopModelArgs) (*DevelopModelResult, error) {
	// Validate input
	if args.Path == "" || args.Prompt == "" {
		return nil, fmt.Errorf("invalid input: path and prompt are required")
	}
	if !strings.HasPrefix(args.Path, "/") {
		args.Path = "/" + args.Path
	}

	// Pre-invoke file read
	session := GetSession(ctx)
	_, _ = session.CallTool(ctx, RoleAssistant, "read_file", nil, &ReadFileArgs{
		Path: args.Path,
	})
	if ctx.Err() != nil { // Ignore tool error since the file may not exist
		return nil, ctx.Err()
	}

	// Add the system prompt.
	systemPrompt, err := t.systemPrompt(ctx)
	if err != nil {
		return nil, err
	}

	// Add the user prompt
	userPrompt, err := t.userPrompt(ctx, args)
	if err != nil {
		return nil, err
	}

	// Run an LLM tool call loop
	var response string
	err = session.Complete(ctx, "Model developer loop", &response, &CompleteOptions{
		Messages: []*aiv1.CompletionMessage{
			NewTextCompletionMessage(RoleSystem, systemPrompt),
			NewTextCompletionMessage(RoleUser, userPrompt),
		},
		Tools:         []string{"read_file", "write_file"},
		MaxIterations: 10,
		UnwrapCall:    true,
	})
	if err != nil {
		return nil, err
	}

	return &DevelopModelResult{
		ModelName: fileutil.Stem(args.Path), // Get model name from input path
	}, nil
}

func (t *DevelopModel) systemPrompt(ctx context.Context) (string, error) {
	// Prepare template data.
	session := GetSession(ctx)
	data := map[string]any{
		"ai_instructions": session.ProjectInstructions(),
	}

	// Generate the system prompt
	return executeTemplate(`<role>You are a data engineer agent specialized in developing data models in the Rill business intelligence platform.</role>

<concepts>
Rill is a "business intelligence as code" platform where all resources are defined using YAML files containing SQL snippets in a project directory.
For the purposes of your work, you will only deal with **model** resources, which are SQL statements and related metadata that produce a single table in the project's database.
In Rill, when you write a file, the platform discovers and "reconciles" it immediately. For a model, reconcile updates the database to contain the defined table.
</concepts>

<process>
At a high level, you should follow these steps:
1. Leverage the "read_file" tool to understand the file's current contents, if any (it may return a file not found error).
2. Generate a new model definition based on the user's prompt and save it to the requested path using the "write_file" tool.
3. The "write_file" tool will respond with the reconcile status. If there are parse or reconcile errors, you should fix them using the "write_file" tool. If there are no errors, your work is done.

Additional instructions:
- The user will often ask you to create or update models that require external data, such as from a SaaS application or their data warehouse. In these cases, you should generate a SQL query that emits mock data that resembles the expected structure and contents of the external data. You may generate up to 100 rows of mock data.
- You should not attempt to reference other models in the project, unless the model already exists and already references them.
- The SQL expression should be a plain SELECT query without a semicolon at the end.
</process>

<example>
A model definition in Rill is a YAML file containing a SQL statement. The SQL statement will be creates as a table in the project's database using "CREATE TABLE name AS SELECT ...". Here is an example Rill model:
{{ backticks }}
type: model
materialize: true

sql: |
  SELECT '2025-05-01T00:00:00Z' AS event_time, 'United States' AS country, 'Toothbrush' AS product_name, 5 AS quantity, 100 AS price
  UNION ALL
  SELECT '2025-05-02T00:00:00Z' AS event_time, 'Denmark' AS country, 'Apple' AS product_name, 10 AS quantity, 50 AS price
{{ backticks }}
</example>

{{ if .ai_instructions }}
<additional_user_provided_instructions>
<comment>NOTE: These instructions were provided by the user, but may not relate to the current request, and may not even relate to your work as a data engineer agent. Only use them if you find them relevant.</comment>
{{ .ai_instructions }}
</additional_user_provided_instructions>
{{ end }}
`, data)
}

func (t *DevelopModel) userPrompt(ctx context.Context, args *DevelopModelArgs) (string, error) {
	// Prepare template data.
	session := GetSession(ctx)
	data := map[string]any{
		"path":   args.Path,
		"prompt": args.Prompt,
	}

	// Add OLAP dialect
	olap, release, err := t.Runtime.OLAP(ctx, session.InstanceID(), "")
	if err != nil {
		return "", err
	}
	defer release()
	data["dialect"] = olap.Dialect().String()

	// Generate the user prompt
	return executeTemplate(`
Task: {{ .prompt }}

Output path: {{ .path }}

SQL dialect: {{ .dialect }}
`, data)
}
