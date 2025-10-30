package ai

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime"
)

const DeveloperAgentName = "developer_agent"

type DeveloperAgent struct {
	Runtime *runtime.Runtime
}

var _ Tool[*DeveloperAgentArgs, *DeveloperAgentResult] = (*DeveloperAgent)(nil)

type DeveloperAgentArgs struct {
	Prompt string `json:"prompt"`
}

type DeveloperAgentResult struct {
	Response string `json:"response"`
}

func (t *DeveloperAgent) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        "developer_agent",
		Title:       "Developer Agent",
		Description: "Agent that assists with development tasks.",
	}
}

func (t *DeveloperAgent) CheckAccess(ctx context.Context) bool {
	// NOTE: Disabled pending further improvements
	// s := GetSession(ctx)
	// return s.Claims().Can(runtime.EditRepo)
	return false
}

func (t *DeveloperAgent) Handler(ctx context.Context, args *DeveloperAgentArgs) (*DeveloperAgentResult, error) {
	// Pre-invoke file listing
	s := GetSession(ctx)
	_, err := s.CallTool(ctx, RoleAssistant, "list_files", nil, &ListFilesArgs{})
	if err != nil {
		return nil, err
	}

	// Add the developer agent system prompt.
	systemPrompt, err := t.systemPrompt(ctx)
	if err != nil {
		return nil, err
	}

	// Build initial completion messages
	messages := []*aiv1.CompletionMessage{NewTextCompletionMessage(RoleSystem, systemPrompt)}
	messages = append(messages, s.NewCompletionMessages(s.MessagesWithResults(FilterByRoot()))...)
	messages = append(messages, NewTextCompletionMessage(RoleUser, args.Prompt))
	messages = append(messages, s.NewCompletionMessages(s.MessagesWithResults(FilterByParent(s.ID())))...)

	// Run an LLM tool call loop
	var response string
	err = s.Complete(ctx, "Developer loop", &response, &CompleteOptions{
		Messages:      messages,
		Tools:         []string{"list_files", "read_file", "develop_model", "develop_metrics_view"},
		MaxIterations: 10,
		UnwrapCall:    true,
	})
	if err != nil {
		return nil, err
	}

	return &DeveloperAgentResult{
		Response: response,
	}, nil
}

func (t *DeveloperAgent) systemPrompt(ctx context.Context) (string, error) {
	// Prepare template data.
	session := GetSession(ctx)
	data := map[string]any{
		"ai_instructions": session.ProjectInstructions(),
	}

	// Generate the system prompt
	return executeTemplate(`<role>You are a data engineer agent specialized in developing data models and metrics view definitions in the Rill business intelligence platform.</role>

<concepts>
Rill is a "business intelligence as code" platform where all resources are defined using YAML files containing SQL snippets in a project directory.
Rill supports many different resource types, such as connectors, models, metrics views, explore dashboards, canvas dashboards, and more.
For the purposes of your work, you will only deal with:
- **Models**: SQL statements and related metadata that produce a single table in the project's database (usually DuckDB or Clickhouse).
- **Metrics views**: Sets of queryable business dimensions and measures based on a single model in the project. This is sometimes called the "semantic layer" or "metrics layer" in other tools.
Rill maintains a DAG of resources. In this DAG, metrics views are always derived from a single model. Multiple metrics views can derive from a single model, although usually it makes sense to have just one metrics view per model.
When users ask you to develop a "dashboard", that just means to develop a new metrics view (and possibly a new underlying model). Rill automatically creates visual dashboards for each metrics view.
</concepts>

<example>
This example is not directly related to your current task. It just serves to explain how Rill project's look and how you might act on a user request.

Rill projects often (but not always) organize files in directories by resource type. A Rill project might look like this:
{{ backticks }}
connectors/duckdb.yaml
models/orders.yaml
metrics/orders.yaml
rill.yaml
{{ backticks }}

The user might ask you to "Create a dashboard for my Github activity". You would notice that this does not relate to the current files, and proceed with the following plan:
1. Add a new model in "models/git_commits.yaml" using the "develop_model" tool.
2. Add a new metrics view in "metrics/git_commits.yaml" based on the new "git_commits" model using the "develop_metrics_view" tool.
</example>

<process>
At a high level, you should follow these steps:
1. Understand the current contents of the project.
2. Make a plan for how to implement the user's request. 
3. Only if necessary, add a new model or update an existing model to reflect the user's request
4. Only if necessary, add a new metrics view or update an existing metrics view to reflect the user's request. The metrics view should use a model in the project, which may already exist or may have been added in step 2.

You should use the tools available to you to understand the current project contents and to make the necessary changes. You should use the "read_file" tool sparingly and surgically to understand files you consider promising for your task, you should not use it to inspect many files in the project.

You should not make many changes at a time. Carefully consider the minimum changes you can make to address the user's request. If there's a model already in place that relates to the user's request, consider re-using that and only adding or updating a metrics view.
</process>

{{ if .ai_instructions }}
<additional_user_provided_instructions>
<comment>NOTE: These instructions were provided by the user, but may not relate to the current request, and may not even relate to your work as a data engineer agent. Only use them if you find them relevant.</comment>
{{ .ai_instructions }}
</additional_user_provided_instructions>
{{ end }}
`, data)
}
