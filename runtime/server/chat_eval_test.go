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
	"time"

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

	t.Run("follows_prescribed_analysis_process", func(t *testing.T) {
		server, instanceID := setupEvalServer(t)

		response := runProjectChatCompletion(t, server, instanceID, "Analyze our advertising performance and find actionable insights")

		// Verify it follows the prescribed 4-phase process from the prompt
		assertContainsToolCall(t, response, "list_metrics_views", "Should discover available datasets")
		assertContainsToolCall(t, response, "get_metrics_view", "Should understand metrics structure")
		assertContainsToolCall(t, response, "query_metrics_view_time_range", "Should scope data availability")
		assertContainsToolCall(t, response, "query_metrics_view", "Should perform analytical queries")

		// Should perform multiple analytical queries (minimum 4-6 per prompt)
		queryCount := countToolCalls(response, "query_metrics_view")
		require.GreaterOrEqual(t, queryCount, 4, "Should perform at least 4 analytical queries")
	})

	t.Run("provides_actionable_insights", func(t *testing.T) {
		server, instanceID := setupEvalServer(t)

		response := runProjectChatCompletion(t, server, instanceID, "What are the key trends in our programmatic advertising performance?")

		// Use LLM-as-a-judge to evaluate business insight quality
		criteria := `
		Evaluate if this response provides actionable business insights:
		1. Identifies specific business problems or opportunities
		2. Recommends concrete actions (not just observations)
		3. Prioritizes recommendations by impact/urgency
		4. Connects findings to business outcomes (revenue, efficiency, etc.)
		
		Rate 1-10 where 10 = excellent actionable business recommendations.`

		responseText := extractResponseText(response)
		score := evaluateWithLLM(t, server, instanceID, responseText, criteria)
		require.GreaterOrEqual(t, score, 7.0, "Response should provide actionable business insights")
	})

	t.Run("demonstrates_data_accuracy", func(t *testing.T) {
		server, instanceID := setupEvalServer(t)

		response := runProjectChatCompletion(t, server, instanceID, "Calculate the average bid price and win rate by advertiser")

		// Use LLM-as-a-judge to verify data accuracy and transparency
		criteria := `
		Evaluate if this response demonstrates data accuracy:
		1. Uses exact numbers from tool results (no rounding errors or approximations)
		2. Shows calculations transparently when needed
		3. Acknowledges data limitations or sample sizes
		4. Avoids making claims beyond what the data supports
		
		Rate 1-10 where 10 = perfect data accuracy and transparency.`

		responseText := extractResponseText(response)
		score := evaluateWithLLM(t, server, instanceID, responseText, criteria)
		require.GreaterOrEqual(t, score, 7.0, "Response should demonstrate data accuracy")
	})

	t.Run("uses_proper_formatting", func(t *testing.T) {
		server, instanceID := setupEvalServer(t)

		response := runProjectChatCompletion(t, server, instanceID, "Summarize the key insights from our programmatic advertising data")

		responseText := extractResponseText(response)

		// Should follow the markdown format specified in the prompt
		require.Contains(t, responseText, "##", "Should use markdown headers for insights")

		// Should have structured content (not just a wall of text)
		require.True(t,
			strings.Contains(responseText, "**") || strings.Contains(responseText, "*") ||
				strings.Contains(responseText, "-") || strings.Contains(responseText, "1."),
			"Should use markdown formatting for structure (bold, lists, etc.)")
	})

	t.Run("suggests_follow_up_analysis", func(t *testing.T) {
		server, instanceID := setupEvalServer(t)

		response := runProjectChatCompletion(t, server, instanceID, "What does this advertising data tell us?")

		responseText := extractResponseText(response)

		// Use LLM-as-a-judge to evaluate follow-up suggestions
		criteria := `
		Evaluate if this response suggests helpful follow-up analysis:
		1. Offers specific next steps or deeper analysis options
		2. Suggests related questions the user might want to explore
		3. Indicates areas that could benefit from further investigation
		4. Helps continue the analytical conversation productively
		
		Rate 1-10 where 10 = excellent follow-up suggestions that guide further analysis.`

		score := evaluateWithLLM(t, server, instanceID, responseText, criteria)
		require.GreaterOrEqual(t, score, 7.0, "Response should suggest helpful follow-up analysis")
	})

	t.Run("rejects_unrelated_questions", func(t *testing.T) {
		server, instanceID := setupEvalServer(t)

		response := runProjectChatCompletion(t, server, instanceID, "Who is the president of India?")

		responseText := extractResponseText(response)

		// Use LLM-as-a-judge to evaluate proper guardrails
		criteria := `
		Evaluate if this response properly handles an unrelated question:
		1. Politely declines to answer the unrelated question
		2. Explains that it's designed for data analysis of the project
		3. Redirects the conversation back to data-related topics
		4. Does NOT attempt to answer the non-data question
		
		Rate 1-10 where 10 = perfectly handles unrelated questions with appropriate guardrails.`

		score := evaluateWithLLM(t, server, instanceID, responseText, criteria)
		require.GreaterOrEqual(t, score, 7.0, "Should politely reject unrelated questions and redirect to data analysis")
	})

	t.Run("uses_rill_comparison_features", func(t *testing.T) {
		server, instanceID := setupEvalServer(t)

		response := runProjectChatCompletion(t, server, instanceID, "How has performance changed over time?")

		// Inspect tool calls to verify the AI uses comparison measures and time ranges
		hasComparisonMeasures := false
		hasComparisonTimeRange := false

		for _, msg := range response {
			for _, block := range msg.Content {
				if toolCall := block.GetToolCall(); toolCall != nil && toolCall.Name == "query_metrics_view" {
					if toolCall.Input == nil {
						continue
					}

					inputMap := toolCall.Input.AsMap()

					// Check for comparison measures in measures array
					if measures, ok := inputMap["measures"].([]interface{}); ok {
						for _, measure := range measures {
							if measureMap, ok := measure.(map[string]interface{}); ok {
								if compute, ok := measureMap["compute"].(map[string]interface{}); ok {
									if _, hasCompDelta := compute["comparison_delta"]; hasCompDelta {
										hasComparisonMeasures = true
									}
									if _, hasCompRatio := compute["comparison_ratio"]; hasCompRatio {
										hasComparisonMeasures = true
									}
								}
								// Also check for delta_abs/delta_rel in measure names
								if name, ok := measureMap["name"].(string); ok {
									if strings.Contains(name, "delta_abs") || strings.Contains(name, "delta_rel") {
										hasComparisonMeasures = true
									}
								}
							}
						}
					}

					// Check for comparison_time_range parameter
					if _, ok := inputMap["comparison_time_range"]; ok {
						hasComparisonTimeRange = true
					}
				}
			}
		}

		// Should use both comparison measures AND comparison_time_range for proper time-based analysis
		require.True(t, hasComparisonMeasures && hasComparisonTimeRange,
			"Should use both comparison measures and comparison_time_range for time analysis")
	})

	t.Run("uses_proper_comparison_time_ranges", func(t *testing.T) {
		server, instanceID := setupEvalServer(t)

		response := runProjectChatCompletion(t, server, instanceID, "Compare this month's performance to last month")

		// Inspect tool calls to validate proper time range usage
		var foundValidComparison bool

		for _, msg := range response {
			for _, block := range msg.Content {
				if toolCall := block.GetToolCall(); toolCall != nil && toolCall.Name == "query_metrics_view" {
					if toolCall.Input == nil {
						continue
					}

					inputMap := toolCall.Input.AsMap()

					// Only validate if this call uses comparison features
					if _, ok := inputMap["comparison_time_range"]; !ok {
						continue
					}

					timeRange, hasTimeRange := inputMap["time_range"].(map[string]interface{})
					comparisonTimeRange, hasComparisonTimeRange := inputMap["comparison_time_range"].(map[string]interface{})

					if !hasTimeRange || !hasComparisonTimeRange {
						continue
					}

					// Extract time range boundaries
					baseStart, _ := timeRange["start"].(string)
					baseEnd, _ := timeRange["end"].(string)
					compStart, _ := comparisonTimeRange["start"].(string)
					compEnd, _ := comparisonTimeRange["end"].(string)

					if baseStart == "" || baseEnd == "" || compStart == "" || compEnd == "" {
						continue
					}

					// Parse times for validation
					baseStartTime, err1 := time.Parse(time.RFC3339, baseStart)
					baseEndTime, err2 := time.Parse(time.RFC3339, baseEnd)
					compStartTime, err3 := time.Parse(time.RFC3339, compStart)
					compEndTime, err4 := time.Parse(time.RFC3339, compEnd)

					if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
						continue
					}

					// Validate time ranges
					baseDuration := baseEndTime.Sub(baseStartTime)
					compDuration := compEndTime.Sub(compStartTime)

					// Check for non-overlapping ranges
					noOverlap := baseEndTime.Before(compStartTime) || compEndTime.Before(baseStartTime) ||
						baseEndTime.Equal(compStartTime) || compEndTime.Equal(baseStartTime)

					// Check for similar duration (within 20% tolerance for practical comparisons)
					durationRatio := float64(baseDuration) / float64(compDuration)
					similarDuration := durationRatio >= 0.8 && durationRatio <= 1.2

					if noOverlap && similarDuration {
						foundValidComparison = true
						break
					}
				}
			}
			if foundValidComparison {
				break
			}
		}

		require.True(t, foundValidComparison,
			"Should use proper non-overlapping time ranges of similar length for comparisons")
	})

	t.Run("acknowledges_query_limitations", func(t *testing.T) {
		server, instanceID := setupEvalServer(t)

		response := runProjectChatCompletion(t, server, instanceID, "Show me performance by publisher")

		// Check if any query_metrics_view calls use limits
		hasLimitedQuery := false
		for _, msg := range response {
			for _, block := range msg.Content {
				if toolCall := block.GetToolCall(); toolCall != nil && toolCall.Name == "query_metrics_view" {
					if toolCall.Input == nil {
						continue
					}
					inputMap := toolCall.Input.AsMap()
					if limit, ok := inputMap["limit"]; ok && limit != nil {
						hasLimitedQuery = true
						break
					}
				}
			}
			if hasLimitedQuery {
				break
			}
		}

		// Only test acknowledgment if the AI actually used limits
		if hasLimitedQuery {
			responseText := extractResponseText(response)
			criteria := `
			Evaluate if this response properly acknowledges query limitations:
			1. Explicitly mentions that results are limited/truncated (e.g., "showing top 10")
			2. Indicates there may be additional data not shown
			3. Explains the impact on analysis conclusions
			4. Offers options to refine or expand the analysis
			
			Rate 1-10 where 10 = perfectly acknowledges limitations and their impact.`

			score := evaluateWithLLM(t, server, instanceID, responseText, criteria)
			require.GreaterOrEqual(t, score, 7.0, "Should acknowledge when using limited queries and explain impact")
		}
	})

	t.Run("handles_high_cardinality_requests_responsibly", func(t *testing.T) {
		server, instanceID := setupEvalServer(t)

		response := runProjectChatCompletion(t, server, instanceID, "Analyze performance across all publishers in detail")

		responseText := extractResponseText(response)

		// Use LLM-as-a-judge to evaluate responsible handling of high-cardinality requests
		criteria := `
		Evaluate if this response responsibly handles a high-cardinality analysis request:
		1. Recognizes that "all publishers" would be too broad/overwhelming
		2. Suggests focused alternatives (e.g., "top performers", "bottom performers", specific segments)
		3. Explains why a focused approach is more valuable than showing everything
		4. Offers concrete next steps for targeted analysis
		
		Rate 1-10 where 10 = perfectly guides user toward focused, actionable analysis.`

		score := evaluateWithLLM(t, server, instanceID, responseText, criteria)
		require.GreaterOrEqual(t, score, 7.0, "Should guide users toward focused analysis instead of overwhelming high-cardinality dumps")
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

	// Automatically log response details for debugging failed tests
	logResponseOnFailure(t, result.Messages)

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

// logResponseOnFailure logs response details immediately for debugging
// TODO: Make this conditional on test failure once we resolve t.Failed() timing issues
func logResponseOnFailure(t *testing.T, response []*runtimev1.Message) {
	t.Logf("=== Response Details for %s ===", t.Name())
	for i, msg := range response {
		t.Logf("Message %d (role: %s):", i, msg.Role)
		for j, block := range msg.Content {
			if text := block.GetText(); text != "" {
				t.Logf("  Block %d (text): %s", j, text)
			}
			if toolCall := block.GetToolCall(); toolCall != nil {
				t.Logf("  Block %d (tool): %s with input: %v", j, toolCall.Name, toolCall.Input)
			}
			if toolResult := block.GetToolResult(); toolResult != nil {
				t.Logf("  Block %d (result): %v", j, toolResult.Content)
			}
		}
	}
	t.Logf("=== End Response Details ===")
}
