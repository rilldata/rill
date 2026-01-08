package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime"
	"go.uber.org/zap"
)

// UserFeedbackToolName is the tool name used for user feedback messages.
const UserFeedbackToolName = "user_feedback"

// UserFeedback implements the user_feedback tool for recording and attributing feedback.
type UserFeedback struct {
	Runtime *runtime.Runtime
}

var _ Tool[*UserFeedbackArgs, *UserFeedbackResult] = (*UserFeedback)(nil)

// UserFeedbackArgs represents the input arguments for user feedback.
type UserFeedbackArgs struct {
	TargetMessageID string   `json:"target_message_id" jsonschema:"The ID of the message being rated."`
	Sentiment       string   `json:"sentiment" jsonschema:"The sentiment of the feedback: positive or negative.,enum=positive,enum=negative"`
	Categories      []string `json:"categories,omitempty" jsonschema:"Feedback categories (only for negative sentiment)."`
	Comment         string   `json:"comment,omitempty" jsonschema:"Optional free-text comment."`
}

// UserFeedbackResult represents the result of recording user feedback.
// Note: Detailed attribution data is stored in the internal completion message for analytics.
type UserFeedbackResult struct {
	Response string `json:"response"`
}

// FeedbackAttributionResult is the structured output type for AI attribution prediction.
// The jsonschema tags constrain the LLM output to valid attribution values.
type FeedbackAttributionResult struct {
	PredictedAttribution string  `json:"predicted_attribution" jsonschema:"The predicted attribution for the issue.,enum=rill,enum=project,enum=user"`
	AttributionReasoning string  `json:"attribution_reasoning" jsonschema:"Explanation of why this attribution was chosen."`
	SuggestedAction      *string `json:"suggested_action,omitempty" jsonschema:"For project or user attribution, a specific action the user can take to get better results."`
}

func (t *UserFeedback) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        UserFeedbackToolName,
		Title:       "User Feedback",
		Description: "Records user feedback on AI responses. For negative feedback, runs attribution to help triage issues.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Recording feedback...",
			"openai/toolInvocation/invoked":  "Recorded feedback",
		},
	}
}

func (t *UserFeedback) CheckAccess(ctx context.Context) (bool, error) {
	// Must be allowed to use AI features
	s := GetSession(ctx)
	if !s.Claims().Can(runtime.UseAI) {
		return false, nil
	}

	// Only allow for rill user agents since it's not useful in MCP contexts.
	if !strings.HasPrefix(s.CatalogSession().UserAgent, "rill") {
		return false, nil
	}
	return true, nil
}

func (t *UserFeedback) Handler(ctx context.Context, args *UserFeedbackArgs) (*UserFeedbackResult, error) {
	s := GetSession(ctx)

	// For positive feedback, return simple acknowledgment
	if args.Sentiment == "positive" {
		return &UserFeedbackResult{
			Response: "Thanks for the positive feedback! I'm glad I could help.",
		}, nil
	}

	// For negative feedback, run attribution
	attribution, err := t.predictAttribution(ctx, s, args)
	if err != nil {
		// Log the error but still return a response
		s.logger.Warn("failed to analyze feedback", zap.String("session_id", s.id), zap.Error(err))
		return &UserFeedbackResult{
			Response: "Thanks for your feedback. I apologize that my response didn't meet your expectations. I'll use this to improve.",
		}, nil
	}

	return &UserFeedbackResult{
		Response: t.generateFeedbackResponse(attribution),
	}, nil
}

// predictAttribution analyzes feedback to determine attribution.
func (t *UserFeedback) predictAttribution(ctx context.Context, s *Session, feedback *UserFeedbackArgs) (*FeedbackAttributionResult, error) {
	// Find the target message that was downvoted
	targetMsg, ok := s.Message(FilterByID(feedback.TargetMessageID))
	if !ok {
		return nil, fmt.Errorf("target message %q not found", feedback.TargetMessageID)
	}

	// Find the user's original prompt (parent call of the target message)
	var originalPrompt string
	if targetMsg.ParentID != "" {
		parentMsg, ok := s.Message(FilterByID(targetMsg.ParentID))
		if ok && parentMsg.Tool == RouterAgentName && parentMsg.Type == MessageTypeCall {
			var args RouterAgentArgs
			if err := json.Unmarshal([]byte(parentMsg.Content), &args); err == nil {
				originalPrompt = args.Prompt
			}
		}
	}

	// Ask the AI to analyze the feedback and determine why the user's expectations were not met
	var attribution FeedbackAttributionResult
	err := s.Complete(ctx, "Feedback attribution", &attribution, &CompleteOptions{
		Messages: []*aiv1.CompletionMessage{
			NewTextCompletionMessage(RoleSystem, t.systemPrompt()),
			NewTextCompletionMessage(RoleUser, t.buildAttributionPrompt(originalPrompt, targetMsg.Content, feedback)),
		},
	})
	if err != nil {
		return nil, err
	}

	return &attribution, nil
}

func (t *UserFeedback) systemPrompt() string {
	return mustExecuteTemplate(`
<role>
You are analyzing feedback on your own AI responses.
Write in first person ("I") when referring to yourself and second person ("you") when referring to the user.
</role>

<categories>
Classify the feedback into one of three categories:
1. "rill" - You (the AI) made an error, or the user is providing product feedback. For example:
		- Made an error in reasoning or misunderstood a clear question
		- Used tools incorrectly or generated an incorrect response
		- User is pushing back on scope limitations (e.g., guardrails about focusing on data analysis)
2. "project" - The data or metadata needed to answer correctly is missing or insufficient. For example:
		- Requested time range is absent from the data
		- Dimensions or measures are missing in the project
		- Descriptions for dimensions or measures are incomplete or unclear
		- Project-level or metrics view-level AI instructions are missing
3. "user" - The user's question was vague, ambiguous, or lacked sufficient context. You responded reasonably given the input.

If you are unsure which category to use, choose "rill" so the Rill team can take a closer look.
</categories>

<output_format>
Write attribution_reasoning as a brief explanation (1-2 sentences) for internal analytics. Be specific about what went wrong.

For "project" and "user" attribution, provide a suggested_action as a complete sentence starting with an action verb addressed to the user (e.g., "Consider adding...", "Try being more specific about..."). This will be shown to the user, so it should be helpful and actionable.
For "rill" attribution, set suggested_action to null (internal errors don't require user action).
</output_format>
`, nil)
}

func (t *UserFeedback) buildAttributionPrompt(originalPrompt, aiResponse string, feedback *UserFeedbackArgs) string {
	// Only include categories if provided
	var categories string
	if len(feedback.Categories) > 0 {
		categories = strings.Join(feedback.Categories, ", ")
	}

	return mustExecuteTemplate(`
Analyze this user feedback on an AI response and determine attribution.

{{ if .originalPrompt }}
User's original question: {{ .originalPrompt }}
{{ end }}
AI's response: {{ .aiResponse }}

{{ if .categories }}
User's feedback categories: {{ .categories }}
{{ end }}
{{ if .comment }}
User's comment: {{ .comment }}
{{ end }}
Determine attribution and provide your analysis.
`, map[string]any{
		"originalPrompt": originalPrompt,
		"aiResponse":     aiResponse,
		"categories":     categories,
		"comment":        feedback.Comment,
	})
}

// generateFeedbackResponse creates a user-visible response based on attribution results.
// Note: attribution_reasoning is stored for analytics but not shown to users.
func (t *UserFeedback) generateFeedbackResponse(attribution *FeedbackAttributionResult) string {
	var response strings.Builder

	response.WriteString("Thanks for your feedback. ")

	switch attribution.PredictedAttribution {
	case "rill":
		// Internal error - generic acknowledgment only (reasoning is for analytics)
		response.WriteString("I made an error in my response. I'll work on improving.")

	case "project":
		// Missing data/config - show suggested action to help user
		if attribution.SuggestedAction != nil && *attribution.SuggestedAction != "" {
			response.WriteString(*attribution.SuggestedAction)
		} else {
			response.WriteString("This may be related to your project's data or configuration.")
		}

	case "user":
		// Vague question - show suggested action to help user get better results
		if attribution.SuggestedAction != nil && *attribution.SuggestedAction != "" {
			response.WriteString(*attribution.SuggestedAction)
		} else {
			response.WriteString("Try being more specific in your question to get better results.")
		}
	}

	return response.String()
}
