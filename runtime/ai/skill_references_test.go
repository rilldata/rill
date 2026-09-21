package ai_test

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	aiv1 "github.com/rilldata/rill/proto/gen/rill/ai/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// turnFunc produces the assistant message the simulated model returns for one completion call.
type turnFunc func(opts *drivers.CompleteOptions) *aiv1.CompletionMessage

// scriptedAIService is a deterministic drivers.AIService for tests. Each Complete call consumes the next scripted
// turn (falling back to a plain "done" reply once turns are exhausted) and records the messages it was given, so
// tests can assert what the model saw.
type scriptedAIService struct {
	turns []turnFunc

	mu     sync.Mutex
	calls  int
	inputs [][]*aiv1.CompletionMessage
}

var _ drivers.AIService = (*scriptedAIService)(nil)

func (s *scriptedAIService) Complete(_ context.Context, opts *drivers.CompleteOptions) (*drivers.CompleteResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.inputs = append(s.inputs, opts.Messages)

	var turn turnFunc
	if s.calls < len(s.turns) {
		turn = s.turns[s.calls]
	}
	s.calls++
	if turn == nil {
		turn = textTurn("done")
	}

	return &drivers.CompleteResult{
		Message:      turn(opts),
		Provider:     "scripted",
		InputTokens:  1,
		OutputTokens: 1,
	}, nil
}

// textTurn makes the model reply with a plain text message (ending the tool loop).
func textTurn(text string) turnFunc {
	return func(_ *drivers.CompleteOptions) *aiv1.CompletionMessage {
		return &aiv1.CompletionMessage{
			Role:    "assistant",
			Content: []*aiv1.ContentBlock{{BlockType: &aiv1.ContentBlock_Text{Text: text}}},
		}
	}
}

// newSkillReferencesSession creates a session on a project with analyst, developer and always-apply skills, backed by the given simulated model.
func newSkillReferencesSession(t *testing.T, script *scriptedAIService) *ai.Session {
	s, _, _ := newSkillReferencesSessionWithRuntime(t, script)
	return s
}

// newSkillReferencesSessionWithRuntime is like newSkillReferencesSession, and also returns the runtime and instance so the test can change the project's files.
func newSkillReferencesSessionWithRuntime(t *testing.T, script *scriptedAIService) (*ai.Session, *runtime.Runtime, string) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": ``,
			"skills/monthly-close/SKILL.md": `---
description: Runs the monthly close analysis.
agents: [analyst]
---
Compare revenue month over month.`,
			"skills/churn-review/SKILL.md": `---
description: Reviews customer churn.
agents: [analyst, developer]
---
List the countries with the most churned customers.`,
			"skills/glossary/SKILL.md": `---
description: Business glossary.
agents: [analyst]
always_apply: true
---
ARPU excludes trial users.`,
			// Without agents, a skill only applies to the developer agent.
			"skills/dev-conventions/SKILL.md": `---
description: Development conventions.
---
Name models in snake_case.`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 5, 0, 0)

	s := newSession(t, rt, instanceID)
	s.SetLLM(func(_ context.Context) (drivers.AIService, func(), error) {
		return script, func() {}, nil
	})
	return s, rt, instanceID
}

// loadedSkillNames returns the names of the skills loaded with load_skill as sub-calls of the given call, in order.
func loadedSkillNames(s *ai.Session, callID string) []string {
	var names []string
	for _, call := range s.Messages(ai.FilterByParent(callID), ai.FilterByType(ai.MessageTypeCall), ai.FilterByTool(ai.LoadSkillName)) {
		names = append(names, s.MustUnmarshalMessageContent(call).(*ai.LoadSkillArgs).Name)
	}
	return names
}

// loadSkillResults returns the tool result content of each load_skill call in the completion messages, by skill name.
func loadSkillResults(messages []*aiv1.CompletionMessage) map[string]string {
	names := map[string]string{} // tool call ID -> skill name
	res := map[string]string{}
	for _, m := range messages {
		for _, block := range m.Content {
			if call := block.GetToolCall(); call != nil && call.Name == ai.LoadSkillName {
				names[call.Id] = call.Input.AsMap()["name"].(string)
			}
			if result := block.GetToolResult(); result != nil {
				if name, ok := names[result.Id]; ok {
					res[name] = result.Content
				}
			}
		}
	}
	return res
}

// TestAnalystLoadsReferencedSkills verifies that the analyst loads the skills referenced with a chat-reference tag in the prompt before the model's first turn,
// once per distinct analyst skill, and ignores references to skills that don't exist or don't apply to the analyst.
func TestAnalystLoadsReferencedSkills(t *testing.T) {
	script := &scriptedAIService{turns: []turnFunc{textTurn("done")}}
	s := newSkillReferencesSession(t, script)

	prompt := `<chat-reference>type="skill" skill="monthly-close"</chat-reference> for March, then ` +
		`<chat-reference>skill="churn-review" type="skill"</chat-reference> and ` +
		`<chat-reference>skill="monthly-close" type="skill"</chat-reference> again. ` +
		`<chat-reference>type="skill" skill="does-not-exist"</chat-reference> ` +
		`<chat-reference>type="skill" skill="dev-conventions"</chat-reference> ` +
		`<chat-reference>type="skill" skill="glossary"</chat-reference> ` +
		`<chat-reference>type="metricsView" metricsView="orders"</chat-reference>`
	res, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, &ai.AnalystAgentArgs{Prompt: prompt})
	require.NoError(t, err)

	// The always-apply glossary is pre-loaded once; the referenced analyst skills are loaded once each, in order of first reference.
	require.Equal(t, []string{"glossary", "monthly-close", "churn-review"}, loadedSkillNames(s, res.Call.ID))

	// The referenced skills' bodies were in the model's input on its first turn.
	require.NotEmpty(t, script.inputs)
	results := loadSkillResults(script.inputs[0])
	require.Contains(t, results["monthly-close"], "Compare revenue month over month.")
	require.Contains(t, results["churn-review"], "List the countries with the most churned customers.")
	require.NotContains(t, results, "dev-conventions")
	require.NotContains(t, results, "does-not-exist")
}

// TestAnalystSkipsSkillsAlreadyLoaded verifies that a skill whose body is already in the conversation is not loaded again when referenced in a later turn.
func TestAnalystSkipsSkillsAlreadyLoaded(t *testing.T) {
	script := &scriptedAIService{turns: []turnFunc{textTurn("first"), textTurn("second")}}
	s := newSkillReferencesSession(t, script)

	prompt := `<chat-reference>type="skill" skill="monthly-close"</chat-reference> for March`
	res1, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, &ai.AnalystAgentArgs{Prompt: prompt})
	require.NoError(t, err)
	require.Equal(t, []string{"glossary", "monthly-close"}, loadedSkillNames(s, res1.Call.ID))

	// The same skill and the always-apply one, referenced again in a later turn.
	prompt = `<chat-reference>type="skill" skill="monthly-close"</chat-reference> and ` +
		`<chat-reference>type="skill" skill="glossary"</chat-reference> for April`
	res2, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, &ai.AnalystAgentArgs{Prompt: prompt})
	require.NoError(t, err)
	require.Empty(t, loadedSkillNames(s, res2.Call.ID))
}

// TestAnalystLoadsSkillAfterFailedLoad verifies that a load_skill call that failed doesn't count as loaded.
// Like the chat, each message opens a new session on the conversation: the skill fails to load in one
// message, its file is fixed, and a later message that references it loads it.
func TestAnalystLoadsSkillAfterFailedLoad(t *testing.T) {
	script := &scriptedAIService{turns: []turnFunc{textTurn("done")}}
	_, rt, instanceID := newSkillReferencesSessionWithRuntime(t, script)

	claims := &runtime.SecurityClaims{UserID: uuid.NewString(), SkipChecks: true}
	runner := ai.NewRunner(rt, activity.NewNoopClient())
	open := func(sessionID string) *ai.Session {
		s, err := runner.Session(t.Context(), &ai.SessionOptions{
			InstanceID: instanceID,
			SessionID:  sessionID,
			Claims:     claims,
			UserAgent:  "rill-evals",
		})
		require.NoError(t, err)
		s.SetLLM(func(_ context.Context) (drivers.AIService, func(), error) {
			return script, func() {}, nil
		})
		return s
	}

	// The skill doesn't exist yet, e.g. while its file has an error, so loading it fails.
	s1 := open("")
	_, err := s1.CallTool(t.Context(), ai.RoleAssistant, ai.LoadSkillName, nil, &ai.LoadSkillArgs{Name: "quarterly-close"})
	require.ErrorContains(t, err, "not found")
	require.NoError(t, s1.Flush(t.Context()))

	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"skills/quarterly-close/SKILL.md": `---
description: Runs the quarterly close analysis.
agents: [analyst]
---
Compare revenue quarter over quarter.`,
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)

	s2 := open(s1.ID())
	t.Cleanup(func() { require.NoError(t, s2.Flush(t.Context())) })
	prompt := `<chat-reference>type="skill" skill="quarterly-close"</chat-reference> for Q3`
	res, err := s2.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, &ai.AnalystAgentArgs{Prompt: prompt})
	require.NoError(t, err)
	require.Contains(t, loadedSkillNames(s2, res.Call.ID), "quarterly-close")
}

// TestAnalystLoadsReferencedSkillsOnEveryTurn verifies that skills referenced in a later turn of the conversation are loaded in that turn.
func TestAnalystLoadsReferencedSkillsOnEveryTurn(t *testing.T) {
	script := &scriptedAIService{turns: []turnFunc{textTurn("first"), textTurn("second")}}
	s := newSkillReferencesSession(t, script)

	res1, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, &ai.AnalystAgentArgs{Prompt: "Hello"})
	require.NoError(t, err)
	require.Equal(t, []string{"glossary"}, loadedSkillNames(s, res1.Call.ID))

	prompt := `<chat-reference>type="skill" skill="monthly-close"</chat-reference> for March`
	res2, err := s.CallTool(t.Context(), ai.RoleUser, ai.AnalystAgentName, nil, &ai.AnalystAgentArgs{Prompt: prompt})
	require.NoError(t, err)
	require.Equal(t, []string{"monthly-close"}, loadedSkillNames(s, res2.Call.ID))

	// The body was in the model's input on the second turn's first completion.
	require.Len(t, script.inputs, 2)
	require.Contains(t, loadSkillResults(script.inputs[1])["monthly-close"], "Compare revenue month over month.")
}
