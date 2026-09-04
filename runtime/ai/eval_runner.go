package ai

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

// evalAnswerPreviewMaxLen bounds the answer preview stored in eval case results.
const evalAnswerPreviewMaxLen = 500

// EvalCaseRunOptions configures a single AI eval case run.
type EvalCaseRunOptions struct {
	InstanceID string
	Claims     *runtime.SecurityClaims
	Case       *runtimev1.AIEvalCase
	// SuiteNotes is suite-level judge guidance, prepended to the case's notes.
	SuiteNotes string
	// SuiteExplore is the suite-level explore context, used when the case doesn't set one.
	SuiteExplore string
	Agent        string
}

// EvalCaseRunResult is the graded outcome of a single AI eval case run.
type EvalCaseRunResult struct {
	Verdict        runtimev1.AIEvalVerdict
	FailureReasons []string
	JudgeReasoning string
	// SessionID points to the AI session holding the case's full transcript.
	SessionID     string
	AnswerPreview string
}

// RunEvalCase asks the configured agent the case's question in a fresh AI session and grades the outcome.
// Infrastructure problems are reported as an ERROR verdict on the result rather than a Go error,
// so callers can distinguish "the agent was wrong" from "the run broke".
func (r *Runner) RunEvalCase(ctx context.Context, opts *EvalCaseRunOptions) (*EvalCaseRunResult, error) {
	session, err := r.Session(ctx, &SessionOptions{
		InstanceID: opts.InstanceID,
		Claims:     opts.Claims,
		// The "rill/eval" user agent keeps eval sessions out of users' chat conversation lists.
		UserAgent: "rill/eval",
	})
	if err != nil {
		return nil, err
	}
	// Flush with a detached context so the transcript survives case timeouts and cancellations.
	defer func() { _ = session.Flush(context.WithoutCancel(ctx)) }()

	res := &EvalCaseRunResult{SessionID: session.ID()}

	if err := session.UpdateTitle(ctx, fmt.Sprintf("Eval: %s", opts.Case.Name)); err != nil {
		return nil, err
	}

	explore := opts.Case.Explore
	if explore == "" {
		explore = opts.SuiteExplore
	}

	var routerRes RouterAgentResult
	_, err = session.CallTool(ctx, RoleUser, RouterAgentName, &routerRes, &RouterAgentArgs{
		Prompt: opts.Case.Question,
		Agent:  opts.Agent,
		AnalystAgentArgs: &AnalystAgentArgs{
			Explore:       explore,
			DisableCharts: true,
		},
	})
	if err != nil {
		res.Verdict = runtimev1.AIEvalVerdict_AI_EVAL_VERDICT_ERROR
		res.FailureReasons = []string{fmt.Sprintf("agent run failed: %s", err.Error())}
		return res, nil
	}

	res.AnswerPreview = truncateEvalAnswer(routerRes.Response)
	r.gradeEvalCase(ctx, session, opts, routerRes.Response, res)
	return res, nil
}

func truncateEvalAnswer(answer string) string {
	runes := []rune(answer)
	if len(runes) <= evalAnswerPreviewMaxLen {
		return answer
	}
	return string(runes[:evalAnswerPreviewMaxLen]) + "…"
}
