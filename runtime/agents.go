package runtime

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/memory"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/model/providers/openai"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ResolverOptions are the options passed to a resolver initializer.
type AgentInitializerOptions struct {
	Runtime     *Runtime
	InstanceID  string
	Runner      *runner.Runner
	ServerTools ServerTools
}

// AgentInitializer is a function that initializes a Agent.
type AgentInitializer func(ctx context.Context, opts *AgentInitializerOptions) (*agent.Agent, error)

// AgentInitializers tracks agent initializers by name.
var AgentInitializers = make(map[string]AgentInitializer)

// RegisterAgentInitializer registers a resolver initializer by name.
func RegisterAgentInitializer(name string, initializer AgentInitializer) {
	if AgentInitializers[name] != nil {
		panic(fmt.Errorf("resolver already registered for name %q", name))
	}
	AgentInitializers[name] = initializer
}

type AgentRunOptions struct {
	Agent       string
	InstanceID  string
	Input       string
	SessionID   string
	ServerTools ServerTools
}

var agentMemory = memory.NewInMemoryStorage()

func (r *Runtime) RunAgent(ctx context.Context, opts *AgentRunOptions) (string, error) {
	initializer, ok := AgentInitializers[opts.Agent]
	if !ok {
		return "", fmt.Errorf("no agent registered with name %q", opts.Agent)
	}

	// Run the agent
	//
	// This runner should come from the openai/anthropic/openllm drivers
	//
	// Something like r.ResolveAIConnector(ctx, opts.InstanceID) should return a AI driver that returns a runner
	// Create OpenAI provider - Get API key from environment
	// Load .env file (fails silently if not found or has errors)
	_ = godotenv.Load()

	// Get RILL_ADMIN_OPENAI_API_KEY key from .env file
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", status.Error(codes.FailedPrecondition, "OPENAI_API_KEY environment variable is not set")
	}

	provider := openai.NewProvider(apiKey)
	provider.SetDefaultModel("gpt-4o")
	provider.WithRateLimit(50, 100000)
	provider.WithRetryConfig(3, 2*time.Second)

	agentRunner := runner.NewRunner().WithDefaultProvider(provider)
	agentRunner.WithMemory(agentMemory)

	// Create a new agent instance
	a, err := initializer(ctx, &AgentInitializerOptions{
		Runtime:     r,
		InstanceID:  opts.InstanceID,
		Runner:      agentRunner,
		ServerTools: opts.ServerTools,
	})
	if err != nil {
		return "", fmt.Errorf("failed to initialize agent %q: %w", opts.Agent, err)
	}

	result, err := agentRunner.Run(ctx, a, &runner.RunOptions{
		Input:     opts.Input,
		MaxTurns:  10,
		SessionID: opts.SessionID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to run agent %q: %w", opts.Agent, err)
	}

	return fmt.Sprintf("%v", result.FinalOutput), nil
}

// ServerTools defines the interface for server tool operations
// This should go away with refactoring server APIs to execute all business logic in dedicated/runtime pkg
type ServerTools interface {
	GenerateMetricsViewFile(ctx context.Context, req *runtimev1.GenerateMetricsViewFileRequest) (*runtimev1.GenerateMetricsViewFileResponse, error)
}
