package ai_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
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
