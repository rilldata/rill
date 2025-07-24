package agents

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/model/providers/openai"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	"github.com/rilldata/rill/runtime/testruntime"
)

func TestNewSyntheticDataAgent(t *testing.T) {
	t.Skip("Skipping synthetic data agent test as it requires a specific runtime setup i.e. file watcher setup")
	ctx := context.Background()
	modelName := "gpt-4o"

	// Create a minimal runtime (mock or real, depending on your test setup)
	rt, instanceID := testruntime.NewInstance(t)

	// Create a provider for OpenAI
	provider := openai.NewProvider("")

	// Configure the provider
	provider.SetDefaultModel("gpt-4o")

	// Customize rate limits if needed (adjust these based on your OpenAI tier)
	provider.WithRateLimit(50, 100000) // 50 requests per minute, 100,000 tokens per minute

	// Configure retry settings
	provider.WithRetryConfig(3, 2*time.Second)

	agent, err := NewSyntheticDataAgent(ctx, instanceID, modelName, rt)
	if err != nil {
		t.Fatalf("NewSyntheticDataAgent failed: %v", err)
	}
	if agent == nil {
		t.Fatal("Expected agent to be non-nil")
	}

	// Create a runner
	r := runner.NewRunner()
	r.WithDefaultProvider(provider)

	// Run the agent
	fmt.Println("\nSending a basic question to the agent...")
	result, err := r.RunSync(agent, &runner.RunOptions{
		Input:    "Generate a sample SQL",
		MaxTurns: 10,
	})
	if err != nil {
		log.Fatalf("Error running agent: %v", err)
	}

	// Print the result
	fmt.Println("\nAgent response:")
	fmt.Println(result.FinalOutput)

	// If there are any responses, display token usage from the last response
	if len(result.RawResponses) > 0 {
		lastResponse := result.RawResponses[len(result.RawResponses)-1]
		if lastResponse.Usage != nil {
			fmt.Printf("\nToken usage: %d total tokens\n", lastResponse.Usage.TotalTokens)
		}
	}
}
