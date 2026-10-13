package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime"
)

const ClickUIName = "click_ui"

// clickUITimeout bounds how long a click_ui call waits for the browser to report the action's outcome.
var clickUITimeout = 30 * time.Second

// ErrUIActionNotPending is returned when no click_ui call is waiting for the reported action.
var ErrUIActionNotPending = errors.New("no pending UI action for this call")

type ClickUI struct{}

var _ Tool[*ClickUIArgs, *ClickUIResult] = (*ClickUI)(nil)

type ClickUIArgs struct {
	ActionID string `json:"action_id" jsonschema:"Exact ID of a currently available UI action. Only use IDs provided in the current UI context."`
}

type ClickUIResult struct{}

func (t *ClickUI) Spec() *mcp.Tool {
	return &mcp.Tool{
		Name:        ClickUIName,
		Title:       "Click UI action",
		Description: "Click a safe action explicitly exposed in the current UI context and wait for the Rill browser UI to finish running it. Only use an exact available action ID.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: boolPtr(false),
			IdempotentHint:  false,
			OpenWorldHint:   boolPtr(false),
			ReadOnlyHint:    false,
		},
		Meta: map[string]any{
			"openai/toolInvocation/invoking": "Clicking UI action...",
			"openai/toolInvocation/invoked":  "UI action requested",
		},
	}
}

func (t *ClickUI) CheckAccess(ctx context.Context) (bool, error) {
	s := GetSession(ctx)
	if !s.Claims().Can(runtime.UseAI) {
		return false, nil
	}

	// Browser actions are only meaningful for the first-party Rill client.
	if !strings.HasPrefix(s.CatalogSession().UserAgent, "rill") {
		return false, nil
	}
	ui, ok := UIContextFromContext(ctx)
	return ok && len(ui.Actions) > 0, nil
}

func (t *ClickUI) Handler(ctx context.Context, args *ClickUIArgs) (*ClickUIResult, error) {
	if args.ActionID == "" {
		return nil, errors.New("action_id is required")
	}

	ui, ok := UIContextFromContext(ctx)
	if !ok {
		return nil, errors.New("no UI actions are available")
	}
	matches := 0
	for _, action := range ui.Actions {
		if action.ID == args.ActionID {
			matches++
		}
	}
	if matches != 1 {
		return nil, errors.New("action_id is not available on the current page")
	}

	// The browser runs the action when it receives this call message,
	// and reports the outcome through Runner.CompleteUIAction.
	s := GetSession(ctx)
	ch := s.runner.uiActions.register(s.ParentID, s.id, s.claims.UserID)
	defer s.runner.uiActions.unregister(s.ParentID)

	select {
	case errMsg := <-ch:
		if errMsg != "" {
			return nil, fmt.Errorf("UI action failed: %s", errMsg)
		}
		return &ClickUIResult{}, nil
	case <-time.After(clickUITimeout):
		return nil, errors.New("the browser did not confirm that the UI action completed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// CompleteUIAction reports the outcome of a UI action the browser ran for a pending click_ui call.
// An empty errMsg means the action succeeded.
func (r *Runner) CompleteUIAction(sessionID, userID, callMessageID, errMsg string) error {
	return r.uiActions.complete(sessionID, userID, callMessageID, errMsg)
}

// uiActionWaiters tracks click_ui calls waiting for the browser to report an action's outcome.
type uiActionWaiters struct {
	mu      sync.Mutex
	waiters map[string]*uiActionWaiter
}

type uiActionWaiter struct {
	sessionID string
	userID    string
	result    chan string
}

func (w *uiActionWaiters) register(callMessageID, sessionID, userID string) <-chan string {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.waiters == nil {
		w.waiters = make(map[string]*uiActionWaiter)
	}
	// Buffered so the browser's report never blocks, even if the call has just timed out.
	ch := make(chan string, 1)
	w.waiters[callMessageID] = &uiActionWaiter{sessionID: sessionID, userID: userID, result: ch}
	return ch
}

func (w *uiActionWaiters) unregister(callMessageID string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.waiters, callMessageID)
}

func (w *uiActionWaiters) complete(sessionID, userID, callMessageID, errMsg string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	waiter, ok := w.waiters[callMessageID]
	// Report a foreign call the same as a missing one, so callers cannot probe other users' calls.
	if !ok || waiter.sessionID != sessionID || waiter.userID != userID {
		return ErrUIActionNotPending
	}
	delete(w.waiters, callMessageID)
	waiter.result <- errMsg
	return nil
}
