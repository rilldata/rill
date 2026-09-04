package ai_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestUserFeedbackPositive(t *testing.T) {
	// Setup empty project and test session
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})
	s := newSession(t, rt, instanceID)

	// Create a real tool call to target with feedback
	var listRes *ai.ListMetricsViewsResult
	callRes, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListMetricsViewsName, &listRes, &ai.ListMetricsViewsArgs{})
	require.NoError(t, err)

	// Invoke the user_feedback tool with positive feedback targeting the tool result
	var res *ai.FeedbackAgentResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.FeedbackAgentName, &res, &ai.FeedbackAgentArgs{
		TargetMessageID: callRes.Result.ID,
		Sentiment:       "positive",
	})

	// Verify the tool result
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, "Thanks for the positive feedback! I'm glad I could help.", res.Response)
}

func TestUserFeedbackAttribution(t *testing.T) {
	// This test requires LLM for attribution prediction
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		AIConnector: "openai",
	})

	cases := []struct {
		name            string
		userPrompt      string
		aiResponse      string
		comment         string
		wantAttribution string // Expected predicted_attribution value
	}{
		{
			name:            "rill_attribution",
			userPrompt:      "What country has the highest revenue?",
			aiResponse:      "Based on the data, the United States has the highest revenue at $1.2 billion.",
			comment:         "This is completely wrong. The data clearly shows China has the highest revenue. You misread the data.",
			wantAttribution: "rill",
		},
		{
			name:            "project_attribution",
			userPrompt:      "What is the revenue for Q4 2024?",
			aiResponse:      "I don't have data for Q4 2024. The available data only covers up to Q2 2024.",
			comment:         "We need Q4 data but it's not in the system yet.",
			wantAttribution: "project",
		},
		{
			name:            "user_attribution",
			userPrompt:      "Show me the thing",
			aiResponse:      "I'm not sure what you're referring to. Could you please clarify what 'thing' you'd like to see?",
			comment:         "I meant the sales report obviously",
			wantAttribution: "user",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newEval(t, rt, instanceID)

			// Create a RouterAgent call message with the user's prompt.
			// We use AddMessage (rather than CallTool) to control the AI response content for testing specific attribution scenarios.
			routerArgs, err := json.Marshal(ai.RouterAgentArgs{Prompt: c.userPrompt})
			require.NoError(t, err)

			userMsg := s.AddMessage(&ai.AddMessageOptions{
				Role:        ai.RoleUser,
				Type:        ai.MessageTypeCall,
				Tool:        ai.RouterAgentName,
				ContentType: ai.MessageContentTypeJSON,
				Content:     string(routerArgs),
			})

			// Create a controlled AI response message to test specific attribution scenarios
			responseMsg := s.WithParent(userMsg.ID).AddMessage(&ai.AddMessageOptions{
				Role:        ai.RoleAssistant,
				Type:        ai.MessageTypeResult,
				Tool:        ai.RouterAgentName,
				ContentType: ai.MessageContentTypeText,
				Content:     c.aiResponse,
			})

			// Test negative feedback targeting the AI response
			var res *ai.FeedbackAgentResult
			_, err = s.CallTool(t.Context(), ai.RoleUser, ai.FeedbackAgentName, &res, &ai.FeedbackAgentArgs{
				TargetMessageID: responseMsg.ID,
				Sentiment:       "negative",
				Comment:         c.comment,
			})
			require.NoError(t, err)
			require.NotNil(t, res)
			require.Contains(t, res.Response, "Thanks for your feedback")

			// Verify attribution is included directly in the result (not as separate messages)
			require.Equal(t, c.wantAttribution, res.PredictedAttribution, "result: %+v", res)
			require.NotEmpty(t, res.AttributionReasoning, "attribution reasoning should not be empty")
		})
	}
}

func TestUserFeedbackAccessDeniedForNonRillUserAgent(t *testing.T) {
	// Setup empty project
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})

	// Create a session with UseAI permission but non-rill user agent
	claims := &runtime.SecurityClaims{
		UserID:      uuid.NewString(),
		SkipChecks:  false,
		Permissions: []runtime.Permission{runtime.UseAI},
	}
	r := ai.NewRunner(rt, activity.NewNoopClient())
	s, err := r.Session(t.Context(), &ai.SessionOptions{
		InstanceID: instanceID,
		Claims:     claims,
		UserAgent:  "mcp-client", // Non-rill user agent
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		err := s.Flush(t.Context())
		require.NoError(t, err)
	})

	// Try to call user_feedback - should fail with access denied
	var res *ai.FeedbackAgentResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.FeedbackAgentName, &res, &ai.FeedbackAgentArgs{
		TargetMessageID: "some-message-id",
		Sentiment:       "positive",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "access denied")
}

func TestUserFeedbackRecordedInCatalog(t *testing.T) {
	// Setup empty project and test session (no LLM configured; attribution failures are tolerated)
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})
	s := newSession(t, rt, instanceID)

	// Create a real tool call to target with feedback
	var listRes *ai.ListMetricsViewsResult
	callRes, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListMetricsViewsName, &listRes, &ai.ListMetricsViewsArgs{})
	require.NoError(t, err)
	targetID := callRes.Result.ID

	catalog, release, err := rt.Catalog(t.Context(), instanceID)
	require.NoError(t, err)
	defer release()

	// Positive feedback is not recorded in the catalog
	var res *ai.FeedbackAgentResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.FeedbackAgentName, &res, &ai.FeedbackAgentArgs{
		TargetMessageID: targetID,
		Sentiment:       "positive",
	})
	require.NoError(t, err)
	rows, err := catalog.FindAIFeedbackForSession(t.Context(), s.CatalogSession().ID)
	require.NoError(t, err)
	require.Len(t, rows, 0)

	// Negative feedback is recorded as an open rating
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.FeedbackAgentName, &res, &ai.FeedbackAgentArgs{
		TargetMessageID: targetID,
		Sentiment:       "negative",
		Categories:      []string{"other"},
		Comment:         "wrong metric",
	})
	require.NoError(t, err)
	rows, err = catalog.FindAIFeedbackForSession(t.Context(), s.CatalogSession().ID)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, drivers.AIFeedbackKindRating, rows[0].Kind)
	require.Equal(t, drivers.AIFeedbackStatusOpen, rows[0].Status)
	require.Equal(t, "negative", rows[0].Sentiment)
	require.Equal(t, []string{"other"}, rows[0].Categories)
	require.Equal(t, "wrong metric", rows[0].Comment)
	require.Equal(t, targetID, rows[0].TargetMessageID)
	require.NotEmpty(t, rows[0].OwnerID)

	// Repeated feedback on the same target updates the open item instead of duplicating it
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.FeedbackAgentName, &res, &ai.FeedbackAgentArgs{
		TargetMessageID: targetID,
		Sentiment:       "negative",
		Comment:         "still wrong",
	})
	require.NoError(t, err)
	rows, err = catalog.FindAIFeedbackForSession(t.Context(), s.CatalogSession().ID)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "still wrong", rows[0].Comment)

	// A review request upgrades the existing item's kind
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.FeedbackAgentName, &res, &ai.FeedbackAgentArgs{
		TargetMessageID: targetID,
		Sentiment:       "negative",
		Comment:         "please review",
		RequestReview:   true,
	})
	require.NoError(t, err)
	rows, err = catalog.FindAIFeedbackForSession(t.Context(), s.CatalogSession().ID)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, drivers.AIFeedbackKindReviewRequest, rows[0].Kind)
	require.Equal(t, "please review", rows[0].Comment)
	require.Contains(t, res.Response, "An admin will review it shortly")

	// A conversation-level review request (no target) creates a separate item
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.FeedbackAgentName, &res, &ai.FeedbackAgentArgs{
		Sentiment:     "negative",
		Comment:       "the whole conversation went wrong",
		RequestReview: true,
	})
	require.NoError(t, err)
	rows, err = catalog.FindAIFeedbackForSession(t.Context(), s.CatalogSession().ID)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	var conversationLevel *drivers.AIFeedback
	for _, f := range rows {
		if f.TargetMessageID == "" {
			conversationLevel = f
		}
	}
	require.NotNil(t, conversationLevel)
	require.Equal(t, drivers.AIFeedbackKindReviewRequest, conversationLevel.Kind)

	// An upvote retracts the same user's open rating on that target, but never an explicit review request
	callRes2, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListMetricsViewsName, &listRes, &ai.ListMetricsViewsArgs{})
	require.NoError(t, err)
	target2 := callRes2.Result.ID
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.FeedbackAgentName, &res, &ai.FeedbackAgentArgs{
		TargetMessageID: target2,
		Sentiment:       "negative",
		Comment:         "changed my mind later",
	})
	require.NoError(t, err)
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.FeedbackAgentName, &res, &ai.FeedbackAgentArgs{
		TargetMessageID: target2,
		Sentiment:       "positive",
	})
	require.NoError(t, err)
	rows, err = catalog.FindAIFeedbackForSession(t.Context(), s.CatalogSession().ID)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	for _, f := range rows {
		switch {
		case f.TargetMessageID == target2:
			require.Equal(t, drivers.AIFeedbackStatusDismissed, f.Status, "the retracted rating should be dismissed")
		case f.TargetMessageID == "":
			require.Equal(t, drivers.AIFeedbackStatusOpen, f.Status, "review requests must not be retracted by upvotes")
		}
	}
}
