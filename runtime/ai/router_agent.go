package ai

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime"
)

// RouterAgent accepts a human prompt and related context, determines which agent is best suited to handling it, and invokes that agent.
// It is usually the entrypoint for processing human completion requests.
type RouterAgent struct {
	Runtime *runtime.Runtime
}

var _ Tool[*RouterAgentArgs, *RouterAgentResult] = (*RouterAgent)(nil)

type RouterAgentArgs struct {
	Prompt  string `json:"prompt"`
	Agent   string `json:"agent,omitempty" jsonschema:"Optional agent to route the request to. If not specified, the system will infer the best agent."`
	Explore string `json:"explore,omitempty" jsonschema:"Optional explore dashboard name. If provided, the exploration will be limited to this dashboard."`
}

type RouterAgentResult struct {
	Response string `json:"response"`
}

func (t *RouterAgent) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        "router_agent",
		Title:       "Router Agent",
		Description: "Agent that routes messages to the appropriate handler agent.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Routing prompt…",
			"openai/toolInvocation/invoked":  "Prompt completed",
		},
	}
}

func (t *RouterAgent) CheckAccess(claims *runtime.SecurityClaims) bool {
	return claims.Can(runtime.UseAI)
}

func (t *RouterAgent) Handler(ctx context.Context, args *RouterAgentArgs) (*RouterAgentResult, error) {
	// TODO: Handle if previous call is still open or awaiting human input

	// Handle title
	s := GetSession(ctx)
	if s.Title() == "" {
		err := s.UpdateTitle(ctx, promptToTitle(args.Prompt))
		if err != nil {
			return nil, err
		}
	}

	// Create a list of candidate agents that the user has access to.
	candidates := []string{"analyst_agent"}
	candidates = slices.DeleteFunc(candidates, func(agent string) bool {
		tool, ok := s.runner.Tools[agent]
		if !ok {
			panic(fmt.Errorf("unknown tool %q", agent))
		}
		if tool.checkAccess != nil {
			return !tool.checkAccess(s.Claims())
		}
		return false
	})

	// Find agent to invoke
	switch {
	// Specific agent requested
	case args.Agent != "":
		if !slices.Contains(candidates, args.Agent) {
			return nil, fmt.Errorf("agent %q not found", args.Agent)
		}
	// No candidates available
	case len(candidates) == 0:
		return nil, fmt.Errorf("no agents available")
	// Only one candidate available
	case len(candidates) == 1:
		args.Agent = candidates[0]
	// Multiple candidates available; choose an agent using the LLM
	default:
		// Build completion messages for agent choice
		messages := []*aiv1.CompletionMessage{NewTextCompletionMessage(RoleSystem, t.systemPrompt(candidates))}
		messages = append(messages, NewCompletionMessages(s.MessagesWithCallResults(s.Messages(FilterByRoot())))...)
		messages = append(messages, NewTextCompletionMessage(RoleUser, args.Prompt))

		// Run agent choice
		var agentChoice struct {
			Agent string `json:"agent"`
		}
		err := s.Complete(ctx, "Agent choice", &agentChoice, &CompleteOptions{
			Messages: messages,
		})
		if err != nil {
			return nil, err
		}

		// Validate the selected agent.
		// NOTE: If we start seeing hallucinations, we may need to add a retry loop with feedback here.
		if !slices.Contains(candidates, agentChoice.Agent) {
			return nil, fmt.Errorf("agent %q not found", agentChoice.Agent)
		}
		args.Agent = agentChoice.Agent
	}

	// Call the selected agent.
	// We always pass "explore" for context, but some agents may not use it.
	var response *AnalystAgentResult // TODO: Don't hard-code to a single agent
	_, err := s.CallTool(ctx, RoleSystem, args.Agent, &response, map[string]any{
		"explore": args.Explore,
	})
	if err != nil {
		return nil, err
	}
	return &RouterAgentResult{Response: response.Response}, nil
}

func (t *RouterAgent) systemPrompt(candidates []string) string {
	return mustExecuteTemplate(`
You are a routing agent that determines which specialized agent should handle a user's request.
You operate in the context of a business intelligence tool that supports data modeling and data exploration, and more.
Your input includes the user's previous messages and responses, as well as the user's latest message, which you are responsible for routing.
Routing guidelines:
- If the user's question relates to developing or changing the data model or dashboards, you should route to the developer.
- If the user's question relates to retrieving specific business metrics, you should route to the analyst.
- If the user asks a general question, you should route to the analyst.
You must answer with a single agent choice and no further explanation. Pick only from this list of available agents:
{{- range .candidates }}
- {{ . }}
{{- end }}
`, map[string]any{
		"candidates": candidates,
	})
}

var whitespaceRegexp = regexp.MustCompile(`\s+`)

func promptToTitle(message string) string {
	title := whitespaceRegexp.ReplaceAllString(message, " ")
	title = strings.TrimSpace(title)
	if len(title) > 50 {
		title = title[:47] + "..."
	}
	if title == "" {
		return "New Conversation"
	}
	return title
}
