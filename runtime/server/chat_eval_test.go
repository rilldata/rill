//go:build evals
// +build evals

package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	// Import OpenAI driver
	_ "github.com/rilldata/rill/runtime/drivers/openai"
)

// TestProjectPromptEvals tests the project-level chat prompt behavior using MCP tools
func TestProjectPromptEvals(t *testing.T) {
	// Load .env file at the repo root (if any)
	_, currentFile, _, _ := goruntime.Caller(0)
	envPath := filepath.Join(currentFile, "..", "..", "..", ".env")
	_, err := os.Stat(envPath)
	if err == nil {
		require.NoError(t, godotenv.Load(envPath))
	}

	apiKey := getOpenAIAPIKey()
	if apiKey == "" {
		t.Skip("OpenAI API key not set (checked RILL_ADMIN_OPENAI_API_KEY and OPENAI_API_KEY), skipping eval tests")
	}

	t.Run("follows_systematic_analysis_process", func(t *testing.T) {
		server, instanceID := setupEvalServer(t)

		response := runProjectChatCompletion(t, server, instanceID, "Analyze the programmatic advertising performance and find actionable insights")

		// Debug: Print the actual response
		for i, msg := range response {
			t.Logf("Response message %d (role: %s): %v", i, msg.Role, msg.Content)
		}

		// Verify it follows the systematic 4-phase process from the prompt
		assertContainsToolCall(t, response, "list_metrics_views", "Should discover available datasets")
		assertContainsToolCall(t, response, "get_metrics_view", "Should understand metrics structure")
		assertContainsToolCall(t, response, "query_metrics_view_time_range", "Should scope data availability")
		assertContainsToolCall(t, response, "query_metrics_view", "Should perform analytical queries")

		// Should perform multiple analytical queries (minimum 4-6 per prompt)
		queryCount := countToolCalls(response, "query_metrics_view")
		require.GreaterOrEqual(t, queryCount, 4, "Should perform at least 4 analytical queries")
	})

	t.Run("provides_quantified_actionable_insights", func(t *testing.T) {
		server, instanceID := setupEvalServer(t)

		response := runProjectChatCompletion(t, server, instanceID, "What are the key trends in our programmatic advertising performance?")

		// Use LLM-as-a-judge to evaluate insight quality
		criteria := `
		Evaluate if this response provides:
		1. Specific, quantified insights (actual numbers from data)
		2. Actionable business recommendations
		3. Clear connection between data patterns and business implications
		4. Surprising or non-obvious findings
		
		Rate 1-10 where 10 = excellent actionable insights with specific numbers.`

		responseText := extractResponseText(response)
		score := evaluateWithLLM(t, server, instanceID, responseText, criteria)
		require.GreaterOrEqual(t, score, 7.0, "Response should provide high-quality actionable insights")
	})

	t.Run("uses_comparison_features_for_time_analysis", func(t *testing.T) {
		server, instanceID := setupEvalServer(t)

		response := runProjectChatCompletion(t, server, instanceID, "How has performance changed over time?")

		// Should use comparison features (delta_abs, delta_rel) for time-based analysis
		responseText := extractResponseText(response)
		require.True(t,
			strings.Contains(responseText, "delta_abs") || strings.Contains(responseText, "delta_rel") ||
				strings.Contains(responseText, "comparison"),
			"Should use comparison features for time analysis")
	})

	t.Run("maintains_analytical_rigor", func(t *testing.T) {
		server, instanceID := setupEvalServer(t)

		response := runProjectChatCompletion(t, server, instanceID, "Calculate the average bid price and win rate by advertiser")

		// Use LLM-as-a-judge to verify analytical rigor
		criteria := `
		Evaluate if this response demonstrates analytical rigor by:
		1. Only using numbers directly from tool results (no manual calculations)
		2. Being precise about data sources and methodology
		3. Acknowledging any limitations or caveats
		4. Building insights systematically from data
		
		Rate 1-10 where 10 = excellent analytical rigor.`

		responseText := extractResponseText(response)
		score := evaluateWithLLM(t, server, instanceID, responseText, criteria)
		require.GreaterOrEqual(t, score, 7.0, "Should demonstrate strong analytical rigor")
	})

	t.Run("formats_output_correctly", func(t *testing.T) {
		server, instanceID := setupEvalServer(t)

		response := runProjectChatCompletion(t, server, instanceID, "Summarize the key insights from our programmatic advertising data")

		responseText := extractResponseText(response)

		// Should follow the markdown format specified in the prompt
		require.Contains(t, responseText, "Based on my systematic analysis", "Should include systematic analysis acknowledgment")
		require.Contains(t, responseText, "##", "Should use markdown headers for insights")

		// Should offer follow-up analysis options
		require.True(t,
			strings.Contains(strings.ToLower(responseText), "follow-up") ||
				strings.Contains(strings.ToLower(responseText), "next") ||
				strings.Contains(strings.ToLower(responseText), "further"),
			"Should offer follow-up analysis options")
	})
}

// setupEvalServer creates a test server with real OpenAI connector and OpenRTB project
func setupEvalServer(t *testing.T) (*Server, string) {
	rt := testruntime.New(t, true)
	ctx := context.Background()

	// Use the OpenRTB test project path
	_, currentFile, _, _ := goruntime.Caller(0)
	projectPath := filepath.Join(currentFile, "..", "..", "testruntime", "testdata", "openrtb")

	// Create instance with OpenAI connector
	inst := &drivers.Instance{
		Environment:      "test",
		OLAPConnector:    "duckdb",
		RepoConnector:    "repo",
		CatalogConnector: "catalog",
		AIConnector:      "openai",
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: map[string]string{"dsn": projectPath},
			},
			{
				Type:   "duckdb",
				Name:   "duckdb",
				Config: map[string]string{"dsn": ":memory:"},
			},
			{
				Type:   "sqlite",
				Name:   "catalog",
				Config: map[string]string{"dsn": fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())},
			},
			{
				Type: "openai",
				Name: "openai",
				Config: map[string]string{
					"api_key": getOpenAIAPIKey(),
					// Let OpenAI driver use defaults (gpt-4o, temperature 0.2)
				},
			},
		},
		Variables: map[string]string{"rill.stage_changes": "false"},
	}

	err := rt.CreateInstance(ctx, inst)
	require.NoError(t, err)

	ctrl, err := rt.Controller(ctx, inst.ID)
	require.NoError(t, err)

	_, err = ctrl.Get(ctx, runtime.GlobalProjectParserName, false)
	require.NoError(t, err)

	err = ctrl.WaitUntilIdle(ctx, false)
	require.NoError(t, err)

	// Create server with MCP integration
	server, err := NewServer(context.Background(), &Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, inst.ID
}

// evalTestCtx provides authentication context for testing
func evalTestCtx() context.Context {
	return auth.WithClaims(context.Background(), auth.NewOpenClaims())
}

// runProjectChatCompletion runs a completion with PROJECT_CHAT context using real MCP tools
func runProjectChatCompletion(t *testing.T, server *Server, instanceID, userMessage string) []*runtimev1.Message {
	appContext := &runtimev1.AppContext{
		ContextType:     runtimev1.AppContextType_APP_CONTEXT_TYPE_PROJECT_CHAT,
		ContextMetadata: &structpb.Struct{},
	}

	// Use the server's Complete method which has real MCP integration
	result, err := server.Complete(evalTestCtx(), &runtimev1.CompleteRequest{
		InstanceId: instanceID,
		AppContext: appContext,
		Messages: []*runtimev1.Message{
			{
				Role: "user",
				Content: []*aiv1.ContentBlock{
					{
						BlockType: &aiv1.ContentBlock_Text{
							Text: userMessage,
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	return result.Messages
}

// evaluateWithLLM uses LLM-as-a-judge to evaluate response quality
func evaluateWithLLM(t *testing.T, server *Server, instanceID, response, criteria string) float64 {
	judgePrompt := fmt.Sprintf(`You are evaluating an AI assistant's response. Please be a strict but fair judge.

EVALUATION CRITERIA:
%s

RESPONSE TO EVALUATE:
%s

Please provide your evaluation as a single number from 1-10, followed by a brief explanation.
Format: "Score: X" where X is your numeric rating.`, criteria, response)

	// Use the server's Complete method for judging (no app context needed for simple evaluation)
	result, err := server.Complete(evalTestCtx(), &runtimev1.CompleteRequest{
		InstanceId: instanceID,
		Messages: []*runtimev1.Message{
			{
				Role: "user",
				Content: []*aiv1.ContentBlock{
					{
						BlockType: &aiv1.ContentBlock_Text{
							Text: judgePrompt,
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	// Extract score from response
	judgeResponse := extractResponseText(result.Messages)
	score := parseScoreFromResponse(judgeResponse)

	t.Logf("LLM Judge Score: %.1f\nJudge Response: %s", score, judgeResponse)

	return score
}

// Helper functions

func getOpenAIAPIKey() string {
	// Check Rill's standard admin API key first
	if key := os.Getenv("RILL_ADMIN_OPENAI_API_KEY"); key != "" {
		return key
	}
	// Fall back to standard OpenAI environment variable
	return os.Getenv("OPENAI_API_KEY")
}

func extractResponseText(messages []*runtimev1.Message) string {
	var parts []string
	for _, msg := range messages {
		if msg.Role == "assistant" {
			for _, block := range msg.Content {
				if text := block.GetText(); text != "" {
					parts = append(parts, text)
				}
			}
		}
	}
	return strings.Join(parts, "\n")
}

func assertContainsToolCall(t *testing.T, messages []*runtimev1.Message, toolName, description string) {
	found := false
	for _, msg := range messages {
		for _, block := range msg.Content {
			if toolCall := block.GetToolCall(); toolCall != nil && toolCall.Name == toolName {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	require.True(t, found, description+" - should contain tool call: "+toolName)
}

func countToolCalls(messages []*runtimev1.Message, toolName string) int {
	count := 0
	for _, msg := range messages {
		for _, block := range msg.Content {
			if toolCall := block.GetToolCall(); toolCall != nil && toolCall.Name == toolName {
				count++
			}
		}
	}
	return count
}

func parseScoreFromResponse(response string) float64 {
	// Look for "Score: X" pattern
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "score:") {
			scoreStr := strings.TrimSpace(strings.TrimPrefix(line, "Score:"))
			scoreStr = strings.TrimPrefix(scoreStr, "score:")
			scoreStr = strings.TrimSpace(scoreStr)

			if score, err := strconv.ParseFloat(scoreStr, 64); err == nil {
				return score
			}
		}
	}

	// Fallback: look for any number in the response
	words := strings.Fields(response)
	for _, word := range words {
		if score, err := strconv.ParseFloat(word, 64); err == nil && score >= 1 && score <= 10 {
			return score
		}
	}

	return 0.0 // Failed to parse
}
