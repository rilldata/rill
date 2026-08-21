package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

// FeedbackAgentName is the name of the feedback agent.
const FeedbackAgentName = "feedback_agent"

// FeedbackAgent handles user feedback on AI responses.
// For negative feedback, it runs attribution to help triage issues.
type FeedbackAgent struct {
	Runtime *runtime.Runtime
}

var _ Tool[*FeedbackAgentArgs, *FeedbackAgentResult] = (*FeedbackAgent)(nil)

// FeedbackAgentArgs represents the input arguments for the feedback agent.
type FeedbackAgentArgs struct {
	TargetMessageID string   `json:"target_message_id,omitempty" jsonschema:"The ID of the message being rated. If empty, the feedback applies to the conversation as a whole."`
	Sentiment       string   `json:"sentiment" jsonschema:"The sentiment of the feedback: positive or negative.,enum=positive,enum=negative"`
	Categories      []string `json:"categories,omitempty" jsonschema:"Feedback categories (only for negative sentiment)."`
	Comment         string   `json:"comment,omitempty" jsonschema:"Optional free-text comment."`
	RequestReview   bool     `json:"request_review,omitempty" jsonschema:"If true, flags the feedback for review by project admins."`
}

// FeedbackAgentResult represents the result of recording user feedback.
// For negative feedback, includes attribution prediction for analytics and LLM context.
type FeedbackAgentResult struct {
	Response             string  `json:"response"`
	PredictedAttribution string  `json:"predicted_attribution,omitempty"`
	AttributionReasoning string  `json:"attribution_reasoning,omitempty"`
	SuggestedAction      *string `json:"suggested_action,omitempty"`
}

func (t *FeedbackAgent) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        FeedbackAgentName,
		Title:       "Feedback Agent",
		Description: "Records user feedback on AI responses. For negative feedback, runs attribution to help triage issues.",
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Recording feedback...",
			"openai/toolInvocation/invoked":  "Recorded feedback",
		},
	}
}

func (t *FeedbackAgent) CheckAccess(ctx context.Context) (bool, error) {
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

func (t *FeedbackAgent) Handler(ctx context.Context, args *FeedbackAgentArgs) (*FeedbackAgentResult, error) {
	s := GetSession(ctx)

	// For positive feedback without a review request, return a simple acknowledgment.
	// Positive ratings are recorded as messages only; they don't need admin attention.
	// If the same user previously downvoted the same target, the upvote retracts that open rating,
	// so a changed mind doesn't leave a stale inbox item that pins the session against TTL cleanup.
	if args.Sentiment == "positive" && !args.RequestReview {
		if err := t.retractOpenRating(ctx, s, args.TargetMessageID); err != nil {
			s.logger.Warn("failed to retract open rating", zap.String("session_id", s.id), zap.Error(err))
		}
		return &FeedbackAgentResult{
			Response: "Thanks for the positive feedback! I'm glad I could help.",
		}, nil
	}

	// Run attribution to pre-triage the item for the review inbox.
	// Attribution failures are tolerated; the feedback is still recorded.
	attribution, err := t.predictAttribution(ctx, s, args)
	if err != nil {
		s.logger.Warn("failed to analyze feedback", zap.String("session_id", s.id), zap.Error(err))
		attribution = nil
	}

	// Record the feedback in the catalog so admins can review it across conversations.
	// Recording failures are logged but don't fail the request; the feedback still exists as messages.
	if err := t.recordFeedback(ctx, s, args, attribution); err != nil {
		s.logger.Error("failed to record feedback", zap.String("session_id", s.id), zap.Error(err))
	}

	res := &FeedbackAgentResult{}
	if attribution != nil {
		res.PredictedAttribution = attribution.PredictedAttribution
		res.AttributionReasoning = attribution.AttributionReasoning
		res.SuggestedAction = attribution.SuggestedAction
	}
	switch {
	case args.RequestReview:
		res.Response = "Feedback submitted. An admin will review it shortly."
	case attribution != nil:
		res.Response = t.generateFeedbackResponse(attribution)
	default:
		res.Response = "Thanks for your feedback. I apologize that my response didn't meet your expectations. I'll use this to improve."
	}
	return res, nil
}

// retractOpenRating dismisses the user's open rating on the given target, if any.
// Explicit review requests are not retracted; they need explicit resolution by an admin.
func (t *FeedbackAgent) retractOpenRating(ctx context.Context, s *Session, targetMessageID string) error {
	catalog, release, err := s.acquireCatalog(ctx)
	if err != nil {
		return err
	}
	defer release()

	items, err := catalog.FindAIFeedbackForSession(ctx, s.id)
	if err != nil {
		return err
	}
	for _, e := range items {
		if e.Status != drivers.AIFeedbackStatusOpen || e.Kind != drivers.AIFeedbackKindRating {
			continue
		}
		if e.TargetMessageID != targetMessageID || e.OwnerID != s.Claims().UserID {
			continue
		}
		now := time.Now()
		e.Status = drivers.AIFeedbackStatusDismissed
		e.ResolvedBy = s.Claims().UserID
		e.ResolvedOn = &now
		if err := catalog.UpdateAIFeedback(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

// recordFeedback inserts (or updates) the feedback item in the catalog's ai_feedback table.
// An existing open item by the same user for the same target is updated in place instead of duplicated;
// a review request upgrades an existing rating's kind.
func (t *FeedbackAgent) recordFeedback(ctx context.Context, s *Session, args *FeedbackAgentArgs, attribution *feedbackAttributionResult) error {
	catalog, release, err := s.acquireCatalog(ctx)
	if err != nil {
		return err
	}
	defer release()

	kind := drivers.AIFeedbackKindRating
	if args.RequestReview {
		kind = drivers.AIFeedbackKindReviewRequest
	}

	f := &drivers.AIFeedback{
		TargetMessageID: args.TargetMessageID,
		Kind:            kind,
		Sentiment:       args.Sentiment,
		Categories:      args.Categories,
		Comment:         args.Comment,
		Status:          drivers.AIFeedbackStatusOpen,
	}
	if attribution != nil {
		f.PredictedAttribution = attribution.PredictedAttribution
		if attribution.SuggestedAction != nil {
			f.SuggestedAction = *attribution.SuggestedAction
		}
	}

	// Update an existing open item by the same user for the same target instead of inserting a duplicate.
	existing, err := catalog.FindAIFeedbackForSession(ctx, s.id)
	if err != nil {
		return err
	}
	for _, e := range existing {
		if e.Status == drivers.AIFeedbackStatusOpen && e.TargetMessageID == args.TargetMessageID && e.OwnerID == s.Claims().UserID {
			e.Sentiment = f.Sentiment
			e.Categories = f.Categories
			e.Comment = f.Comment
			e.PredictedAttribution = f.PredictedAttribution
			e.SuggestedAction = f.SuggestedAction
			if args.RequestReview {
				e.Kind = drivers.AIFeedbackKindReviewRequest
			}
			return catalog.UpdateAIFeedback(ctx, e)
		}
	}

	email, _ := s.Claims().UserAttributes["email"].(string)
	now := time.Now()
	f.ID = uuid.NewString()
	f.InstanceID = s.instanceID
	f.SessionID = s.id
	f.OwnerID = s.Claims().UserID
	f.OwnerEmail = email
	f.CreatedOn = now
	f.UpdatedOn = now
	return catalog.InsertAIFeedback(ctx, f)
}

// feedbackAttributionResult is the structured output type for AI attribution prediction.
// The jsonschema tags constrain the LLM output to valid attribution values.
type feedbackAttributionResult struct {
	PredictedAttribution string  `json:"predicted_attribution" jsonschema:"The predicted attribution for the issue.,enum=rill,enum=project,enum=user"`
	AttributionReasoning string  `json:"attribution_reasoning" jsonschema:"Explanation of why this attribution was chosen."`
	SuggestedAction      *string `json:"suggested_action,omitempty" jsonschema:"For project or user attribution, a specific action the user can take to get better results."`
}

// predictAttribution analyzes feedback to determine attribution.
func (t *FeedbackAgent) predictAttribution(ctx context.Context, s *Session, feedback *FeedbackAgentArgs) (*feedbackAttributionResult, error) {
	// Find the target message that was downvoted.
	// For conversation-level feedback (no target), attribute against the latest agent response.
	var targetMsg *Message
	aiResponse := ""
	if feedback.TargetMessageID != "" {
		msg, ok := s.Message(FilterByID(feedback.TargetMessageID))
		if !ok {
			return nil, fmt.Errorf("target message %q not found", feedback.TargetMessageID)
		}
		targetMsg = msg
		aiResponse = msg.Content
	} else {
		msg, ok := s.LatestMessage(FilterByTool(RouterAgentName), FilterByType(MessageTypeResult))
		if !ok {
			return nil, fmt.Errorf("no agent response found to attribute conversation-level feedback against")
		}
		targetMsg = msg
		// The router result content is JSON; unwrap to the response text when possible.
		aiResponse = msg.Content
		if content, err := s.UnmarshalMessageContent(msg); err == nil {
			if res, ok := content.(*RouterAgentResult); ok {
				aiResponse = res.Response
			}
		}
	}

	// Find the user's original prompt (parent call of the target message)
	var originalPrompt string
	parentMsg, ok := s.Message(FilterByID(targetMsg.ParentID))
	if ok {
		parentContent, err := s.UnmarshalMessageContent(parentMsg)
		if err != nil {
			return nil, err
		}
		if args, ok := parentContent.(*RouterAgentArgs); ok {
			originalPrompt = args.Prompt
		}
	}

	// Ask the AI to analyze the feedback and determine why the user's expectations were not met.
	// UnwrapCall keeps attribution as an internal implementation detail - no separate messages are created.
	// The attribution result is stored in the user_feedback result for analytics and LLM context.
	var attribution feedbackAttributionResult
	err := s.Complete(ctx, "Feedback attribution", &attribution, &CompleteOptions{
		Messages: []*aiv1.CompletionMessage{
			NewTextCompletionMessage(RoleSystem, t.systemPrompt()),
			NewTextCompletionMessage(RoleUser, t.buildAttributionPrompt(originalPrompt, aiResponse, feedback)),
		},
		UnwrapCall: true,
	})
	if err != nil {
		return nil, err
	}

	return &attribution, nil
}

func (t *FeedbackAgent) systemPrompt() string {
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

func (t *FeedbackAgent) buildAttributionPrompt(originalPrompt, aiResponse string, feedback *FeedbackAgentArgs) string {
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
func (t *FeedbackAgent) generateFeedbackResponse(attribution *feedbackAttributionResult) string {
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
