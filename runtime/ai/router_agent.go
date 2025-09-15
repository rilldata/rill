package ai

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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
	Agent   string `json:"agent"`
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
	}
}

func (t *RouterAgent) CheckAccess(claims *runtime.SecurityClaims) bool {
	return true
}

func (t *RouterAgent) Handler(ctx context.Context, args *RouterAgentArgs) (*RouterAgentResult, error) {
	// TODO: Handle if previous call is still open or awaiting human input

	// Handle title
	session := GetSession(ctx)
	if session.Title() == "" {
		session.UpdateTitle(ctx, promptToTitle(args.Prompt))
	}

	// Add prompt to session
	session.AddMessage(&AddMessageOptions{
		Role:        RoleUser,
		Type:        MessageTypePrompt,
		ContentType: MessageContentTypeText,
		Content:     args.Prompt,
	})

	// Create a list of candidate agents that the user has access to.
	candidates := []string{"developer_agent", "analyst_agent"}
	candidates = slices.DeleteFunc(candidates, func(agent string) bool {
		tool, ok := session.runner.Tools[agent]
		if !ok {
			panic(fmt.Errorf("unknown tool %q", agent))
		}
		if tool.checkAccess != nil {
			return !tool.checkAccess(session.Claims())
		}
		return false
	})

	// Handle if a specific agent was requested.
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no agents available")
	}
	if args.Agent != "" {
		if !slices.Contains(candidates, args.Agent) {
			return nil, fmt.Errorf("agent %q not found", args.Agent)
		}
	} else if len(candidates) == 1 {
		args.Agent = candidates[0]
	} else {
		session.AddMessage(&AddMessageOptions{
			Role:        RoleSystem,
			Type:        MessageTypePrompt,
			ContentType: MessageContentTypeText,
			Content:     t.systemPrompt(candidates),
		})
		var agentChoice struct {
			Agent string `json:"agent"`
		}
		err := session.Complete(ctx, "Agent choice", &agentChoice, &CompleteOptions{
			Messages: session.DefaultCompletionMessages(),
		})
		if err != nil {
			return nil, err
		}
		if !slices.Contains(candidates, agentChoice.Agent) {
			return nil, fmt.Errorf("agent %q not found", agentChoice.Agent)
		}
		args.Agent = agentChoice.Agent
	}

	// Call the selected agent.
	// We always pass "explore" for context, but some agents may not use it.
	var response *AnalystAgentResult
	_, err := session.CallTool(ctx, RoleSystem, args.Agent, &response, map[string]any{
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
You operate in the context of a business intelligence tool that supports data modeling, data exploration, dashboarding, and more.
Your input includes the user's previous messages and responses, as well as the user's latest message, which you are responsible for routing.
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
