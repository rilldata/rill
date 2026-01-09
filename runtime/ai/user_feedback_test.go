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

	// Seed a message to target with feedback
	userMsg := s.AddMessage(&ai.AddMessageOptions{
		Role:    ai.RoleUser,
		Type:    ai.MessageTypeCall,
		Tool:    ai.RouterAgentName,
		Content: "What is 2 + 2?",
	})
	responseMsg := s.WithParent(userMsg.ID).AddMessage(&ai.AddMessageOptions{
		Role:    ai.RoleAssistant,
		Type:    ai.MessageTypeResult,
		Tool:    ai.RouterAgentName,
		Content: "2 + 2 = 4",
	})

	// Invoke the user_feedback tool with positive feedback
	var res *ai.UserFeedbackResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.UserFeedbackToolName, &res, &ai.UserFeedbackArgs{
		TargetMessageID: responseMsg.ID,
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
		EnableLLM: true,
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

			// Seed a user message
			routerArgs, err := json.Marshal(ai.RouterAgentArgs{Prompt: c.userPrompt})
			require.NoError(t, err)

			userMsg := s.AddMessage(&ai.AddMessageOptions{
				Role:    ai.RoleUser,
				Type:    ai.MessageTypeCall,
				Tool:    ai.RouterAgentName,
				Content: string(routerArgs),
			})

			// Seed an AI response message
			responseMsg := s.WithParent(userMsg.ID).AddMessage(&ai.AddMessageOptions{
				Role:    ai.RoleAssistant,
				Type:    ai.MessageTypeResult,
				Tool:    ai.RouterAgentName,
				Content: c.aiResponse,
			})

			// Test negative feedback targeting the AI response
			var res *ai.UserFeedbackResult
			callResult, err := s.CallTool(t.Context(), ai.RoleUser, ai.UserFeedbackToolName, &res, &ai.UserFeedbackArgs{
				TargetMessageID: responseMsg.ID,
				Sentiment:       "negative",
				Comment:         c.comment,
			})
			require.NoError(t, err)
			require.NotNil(t, res)
			require.Contains(t, res.Response, "Thanks for your feedback")

			// Find and verify the attribution result from nested completion
			var attribution ai.FeedbackAttributionResult
			for _, msg := range s.Messages(ai.FilterByParent(callResult.Call.ID)) {
				if msg.Tool == "Feedback attribution" && msg.Type == ai.MessageTypeResult {
					err := json.Unmarshal([]byte(msg.Content), &attribution)
					require.NoError(t, err)
					break
				}
			}
			require.Equal(t, c.wantAttribution, attribution.PredictedAttribution, "attribution: %+v", attribution)
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
	var res *ai.UserFeedbackResult
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.UserFeedbackToolName, &res, &ai.UserFeedbackArgs{
		TargetMessageID: "some-message-id",
		Sentiment:       "positive",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "access denied")
}
