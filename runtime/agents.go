package runtime

import (
	"context"
	"fmt"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/memory"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
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

	// Acquire AI
	ai, release, err := r.AI(ctx, opts.InstanceID)
	if err != nil {
		return "", err
	}
	defer release()

	provider := ai.LLMProvider()
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
