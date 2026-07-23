package reconcilers_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	alertTransitionResolverName = "alert_transition_test"
	alertTransitionDriverName   = "alert_transition_test"
	alertTransitionAdminName    = "alert_transition_admin"
	alertTransitionNotifierName = "alert_transition_notifier"
)

var (
	alertTransitionScenarios sync.Map
	alertTransitionSequence  atomic.Uint64
)

func init() {
	runtime.RegisterResolverInitializer(alertTransitionResolverName, newAlertTransitionResolver)
	drivers.Register(alertTransitionDriverName, &alertTransitionDriver{})
}

func TestAlertResolverFailuresCommitExecutionState(t *testing.T) {
	// Resolver setup and iteration errors are alert outcomes: each must finish
	// one execution, advance scheduling, and clear the one-shot trigger.
	tests := []struct {
		name         string
		configure    func(*alertTransitionScenario)
		wantError    string
		wantValidate int
		wantResolve  int
		wantNext     int
	}{
		{
			name: "initializer failure",
			configure: func(s *alertTransitionScenario) {
				s.initializerErr = errors.New("initializer unavailable")
			},
			wantError: "initializer unavailable",
		},
		{
			name: "validation failure",
			configure: func(s *alertTransitionScenario) {
				s.validateErr = errors.New("resolver properties invalid")
			},
			wantError:    "resolver properties invalid",
			wantValidate: 1,
		},
		{
			name: "next failure",
			configure: func(s *alertTransitionScenario) {
				s.nextErr = errors.New("row stream interrupted")
			},
			wantError:    "failed to get row from alert resolver: row stream interrupted",
			wantValidate: 1,
			wantResolve:  1,
			wantNext:     1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Phase-specific call counts prove the failure is handled at the
			// intended resolver boundary instead of by an earlier shortcut.
			scenario, scenarioID := newAlertTransitionScenario(t)
			tt.configure(scenario)
			fixture := newAlertTransitionFixture(t, scenarioID, nil, alertTransitionSpec(t, scenarioID), false, false)

			res := requireAlertCompleted(t, fixture, 1)
			execution := res.GetAlert().State.ExecutionHistory[0]
			require.Equal(t, runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR, execution.Result.Status)
			require.ErrorContains(t, errors.New(execution.Result.ErrorMessage), tt.wantError)
			require.NotNil(t, execution.ExecutionTime)
			require.Empty(t, res.Meta.ReconcileError)

			snapshot := scenario.snapshot()
			require.Equal(t, 1, snapshot.initializerCalls)
			require.Equal(t, tt.wantValidate, snapshot.validateCalls)
			require.Equal(t, tt.wantResolve, snapshot.resolveCalls)
			require.Equal(t, tt.wantNext, snapshot.nextCalls)
		})
	}
}

func TestAlertAdminMetadataFailureCommitsErrorHistory(t *testing.T) {
	// Admin failure happens before query initialization, but it is still a
	// completed alert attempt and must be visible in history and reconcile state.
	scenario, scenarioID := newAlertTransitionScenario(t)
	scenario.adminErr = errors.New("admin metadata unavailable")
	fixture := newAlertTransitionFixture(t, scenarioID, nil, alertTransitionSpec(t, scenarioID), true, false)

	res := requireAlertCompleted(t, fixture, 1)
	execution := res.GetAlert().State.ExecutionHistory[0]
	require.Equal(t, runtimev1.AssertionStatus_ASSERTION_STATUS_ERROR, execution.Result.Status)
	require.ErrorContains(t, errors.New(execution.Result.ErrorMessage), "failed to get alert metadata: admin metadata unavailable")
	require.Nil(t, execution.ExecutionTime)
	require.ErrorContains(t, errors.New(res.Meta.ReconcileError), "failed to get alert metadata")

	snapshot := scenario.snapshot()
	require.Equal(t, 1, snapshot.adminCalls)
	require.Zero(t, snapshot.initializerCalls)
}

func TestAlertEmailFailureRetriesOnlyWhenNoDeliverySucceeded(t *testing.T) {
	// A failed first recipient leaves SentNotifications false, allowing the next
	// explicit execution to retry without manufacturing an assertion transition.
	scenario, scenarioID := newAlertTransitionScenario(t)
	scenario.rows = []map[string]any{{"orders": 1}}
	sender := &alertTransitionEmailSender{errs: []error{errors.New("smtp unavailable"), nil}}
	spec := alertTransitionSpec(t, scenarioID)
	spec.NotifyOnFail = true
	spec.Notifiers = []*runtimev1.Notifier{alertTransitionEmailNotifier(t, "ops@example.com")}
	fixture := newAlertTransitionFixture(t, scenarioID, sender, spec, false, false)

	res := requireAlertCompleted(t, fixture, 1)
	first := res.GetAlert().State.ExecutionHistory[0]
	require.Equal(t, runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, first.Result.Status)
	require.ErrorContains(t, errors.New(first.Result.ErrorMessage), "smtp unavailable")
	require.False(t, first.SentNotifications)
	require.Equal(t, []string{"ops@example.com"}, sender.snapshot())

	triggerAlertAgain(t, fixture)
	res = requireAlertCompleted(t, fixture, 2)
	second := res.GetAlert().State.ExecutionHistory[0]
	require.Equal(t, runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, second.Result.Status)
	require.Empty(t, second.Result.ErrorMessage)
	require.True(t, second.SentNotifications)
	require.Equal(t, []string{"ops@example.com", "ops@example.com"}, sender.snapshot())
}

func TestAlertPartialMultiNotifierFailureDoesNotDuplicateSuccess(t *testing.T) {
	// Email succeeds before the non-email notifier fails. The dirty marker must
	// suppress both transports on the next unchanged failure.
	scenario, scenarioID := newAlertTransitionScenario(t)
	scenario.rows = []map[string]any{{"orders": 1}}
	scenario.notifierErrs = []error{errors.New("webhook unavailable")}
	sender := &alertTransitionEmailSender{}
	spec := alertTransitionSpec(t, scenarioID)
	spec.NotifyOnFail = true
	spec.Notifiers = []*runtimev1.Notifier{
		alertTransitionEmailNotifier(t, "ops@example.com"),
		{Connector: alertTransitionNotifierName, Properties: alertTransitionProperties(t, map[string]any{})},
	}
	fixture := newAlertTransitionFixture(t, scenarioID, sender, spec, false, true)

	res := requireAlertCompleted(t, fixture, 1)
	first := res.GetAlert().State.ExecutionHistory[0]
	require.Equal(t, runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, first.Result.Status)
	require.ErrorContains(t, errors.New(first.Result.ErrorMessage), "webhook unavailable")
	require.True(t, first.SentNotifications)
	require.Equal(t, []string{"ops@example.com"}, sender.snapshot())
	require.Equal(t, 1, scenario.snapshot().notifierCalls)

	triggerAlertAgain(t, fixture)
	res = requireAlertCompleted(t, fixture, 2)
	second := res.GetAlert().State.ExecutionHistory[0]
	require.Equal(t, runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, second.Result.Status)
	require.False(t, second.SentNotifications)
	require.Empty(t, second.Result.ErrorMessage)
	require.Equal(t, []string{"ops@example.com"}, sender.snapshot())
	require.Equal(t, 1, scenario.snapshot().notifierCalls)
}

func TestAlertContextInterruptionBeforeDeliveryRemainsRetryable(t *testing.T) {
	// Clean interruption should preserve the ad-hoc trigger and empty history so
	// a later controller invocation can complete exactly one delivery.
	tests := []struct {
		name string
		err  error
	}{
		{name: "cancellation", err: context.Canceled},
		{name: "deadline", err: context.DeadlineExceeded},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// With no external side effect, cancellation and deadline both leave
			// the execution clean, keep Trigger set, and retry exactly once later.
			scenario, scenarioID := newAlertTransitionScenario(t)
			scenario.resolveErr = tt.err
			sender := &alertTransitionEmailSender{}
			spec := alertTransitionSpec(t, scenarioID)
			spec.NotifyOnFail = true
			spec.Notifiers = []*runtimev1.Notifier{alertTransitionEmailNotifier(t, "ops@example.com")}
			fixture := newAlertTransitionFixture(t, scenarioID, sender, spec, false, false)

			res := requireAlertRetryableInterruption(t, fixture)
			require.Contains(t, res.Meta.ReconcileError, tt.err.Error())
			require.Empty(t, sender.snapshot())

			scenario.setResolveOutcome(nil, []map[string]any{{"orders": 1}})
			reconcileAlert(t, fixture)
			res = requireAlertCompleted(t, fixture, 1)
			require.True(t, res.GetAlert().State.ExecutionHistory[0].SentNotifications)
			require.Equal(t, []string{"ops@example.com"}, sender.snapshot())
		})
	}
}

func TestAlertContextInterruptionAfterDeliveryDoesNotRetry(t *testing.T) {
	// Dirty interruption should commit history and scheduling so reconciliation
	// cannot automatically replay an already-successful recipient.
	tests := []struct {
		name string
		err  error
	}{
		{name: "cancellation", err: context.Canceled},
		{name: "deadline", err: context.DeadlineExceeded},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The first recipient succeeds and the second reports interruption;
			// the committed dirty execution must suppress a duplicate first send.
			scenario, scenarioID := newAlertTransitionScenario(t)
			scenario.rows = []map[string]any{{"orders": 1}}
			sender := &alertTransitionEmailSender{errs: []error{nil, tt.err}}
			spec := alertTransitionSpec(t, scenarioID)
			spec.NotifyOnFail = true
			spec.Notifiers = []*runtimev1.Notifier{alertTransitionEmailNotifier(t, "first@example.com", "second@example.com")}
			fixture := newAlertTransitionFixture(t, scenarioID, sender, spec, false, false)

			res := requireAlertCompleted(t, fixture, 1)
			first := res.GetAlert().State.ExecutionHistory[0]
			require.Equal(t, runtimev1.AssertionStatus_ASSERTION_STATUS_FAIL, first.Result.Status)
			require.Contains(t, first.Result.ErrorMessage, tt.err.Error())
			require.True(t, first.SentNotifications)
			require.Equal(t, []string{"first@example.com", "second@example.com"}, sender.snapshot())

			triggerAlertAgain(t, fixture)
			res = requireAlertCompleted(t, fixture, 2)
			require.False(t, res.GetAlert().State.ExecutionHistory[0].SentNotifications)
			require.Equal(t, []string{"first@example.com", "second@example.com"}, sender.snapshot())
		})
	}
}

type alertTransitionFixture struct {
	rt   *runtime.Runtime
	id   string
	ctrl *runtime.Controller
	name *runtimev1.ResourceName
}

func newAlertTransitionFixture(t *testing.T, scenarioID string, sender email.Sender, spec *runtimev1.AlertSpec, useAdmin, useNotifier bool) *alertTransitionFixture {
	t.Helper()
	rt, id := testruntime.NewInstance(t)
	if sender != nil {
		rt.Email = email.New(sender)
	}

	inst, err := rt.Instance(t.Context(), id)
	require.NoError(t, err)
	edited := *inst
	edited.Connectors = append([]*runtimev1.Connector(nil), inst.Connectors...)
	properties := alertTransitionProperties(t, map[string]any{"scenario": scenarioID})
	if useAdmin {
		edited.AdminConnector = alertTransitionAdminName
		edited.Connectors = append(edited.Connectors, &runtimev1.Connector{Name: alertTransitionAdminName, Type: alertTransitionDriverName, Config: properties})
	}
	if useNotifier {
		edited.Connectors = append(edited.Connectors, &runtimev1.Connector{Name: alertTransitionNotifierName, Type: alertTransitionDriverName, Config: properties})
	}
	require.NoError(t, rt.EditInstance(t.Context(), &edited, false))

	ctrl, err := rt.Controller(t.Context(), id)
	require.NoError(t, err)
	name := &runtimev1.ResourceName{Kind: runtime.ResourceKindAlert, Name: "transition"}
	require.NoError(t, ctrl.Create(t.Context(), name, nil, nil, nil, nil, false, &runtimev1.Resource{
		Resource: &runtimev1.Resource_Alert{Alert: &runtimev1.Alert{Spec: spec, State: &runtimev1.AlertState{}}},
	}))
	require.NoError(t, ctrl.WaitUntilIdle(t.Context(), false))
	return &alertTransitionFixture{rt: rt, id: id, ctrl: ctrl, name: name}
}

func alertTransitionSpec(t *testing.T, scenarioID string) *runtimev1.AlertSpec {
	t.Helper()
	return &runtimev1.AlertSpec{
		DisplayName:        "Transition test",
		Trigger:            true,
		RefreshSchedule:    &runtimev1.Schedule{TickerSeconds: 3600},
		Resolver:           alertTransitionResolverName,
		ResolverProperties: alertTransitionProperties(t, map[string]any{"scenario": scenarioID}),
	}
}

func alertTransitionEmailNotifier(t *testing.T, recipients ...string) *runtimev1.Notifier {
	t.Helper()
	values := make([]any, len(recipients))
	for i, recipient := range recipients {
		values[i] = recipient
	}
	return &runtimev1.Notifier{
		Connector:  "email",
		Properties: alertTransitionProperties(t, map[string]any{"recipients": values}),
	}
}

func alertTransitionProperties(t *testing.T, values map[string]any) *structpb.Struct {
	t.Helper()
	properties, err := structpb.NewStruct(values)
	require.NoError(t, err)
	return properties
}

func requireAlertCompleted(t *testing.T, fixture *alertTransitionFixture, count uint32) *runtimev1.Resource {
	t.Helper()
	res, err := fixture.ctrl.Get(t.Context(), fixture.name, true)
	require.NoError(t, err)
	alert := res.GetAlert()
	require.Nil(t, alert.State.CurrentExecution)
	require.Equal(t, count, alert.State.ExecutionCount)
	require.Len(t, alert.State.ExecutionHistory, int(count))
	require.NotNil(t, alert.State.NextRunOn)
	require.True(t, alert.State.NextRunOn.AsTime().After(time.Now()))
	require.False(t, alert.Spec.Trigger)
	require.NotNil(t, res.Meta.ReconcileOn)
	require.Equal(t, alert.State.NextRunOn.AsTime(), res.Meta.ReconcileOn.AsTime())
	return res
}

func requireAlertRetryableInterruption(t *testing.T, fixture *alertTransitionFixture) *runtimev1.Resource {
	t.Helper()
	res, err := fixture.ctrl.Get(t.Context(), fixture.name, true)
	require.NoError(t, err)
	alert := res.GetAlert()
	require.Nil(t, alert.State.CurrentExecution)
	require.Zero(t, alert.State.ExecutionCount)
	require.Empty(t, alert.State.ExecutionHistory)
	require.Nil(t, alert.State.NextRunOn)
	require.True(t, alert.Spec.Trigger)
	return res
}

func triggerAlertAgain(t *testing.T, fixture *alertTransitionFixture) {
	t.Helper()
	res, err := fixture.ctrl.Get(t.Context(), fixture.name, true)
	require.NoError(t, err)
	res.GetAlert().Spec.Trigger = true
	require.NoError(t, fixture.ctrl.UpdateSpec(t.Context(), fixture.name, res))
	require.NoError(t, fixture.ctrl.WaitUntilIdle(t.Context(), false))
}

func reconcileAlert(t *testing.T, fixture *alertTransitionFixture) {
	t.Helper()
	require.NoError(t, fixture.ctrl.Reconcile(t.Context(), fixture.name))
	require.NoError(t, fixture.ctrl.WaitUntilIdle(t.Context(), false))
}

type alertTransitionScenario struct {
	mu sync.Mutex

	initializerErr error
	validateErr    error
	resolveErr     error
	nextErr        error
	rows           []map[string]any
	adminErr       error
	adminMeta      *drivers.AlertMetadata
	notifierErrs   []error

	initializerCalls int
	validateCalls    int
	resolveCalls     int
	nextCalls        int
	adminCalls       int
	notifierCalls    int
}

type alertTransitionSnapshot struct {
	initializerCalls int
	validateCalls    int
	resolveCalls     int
	nextCalls        int
	adminCalls       int
	notifierCalls    int
}

func newAlertTransitionScenario(t *testing.T) (*alertTransitionScenario, string) {
	t.Helper()
	id := fmt.Sprintf("scenario-%d", alertTransitionSequence.Add(1))
	scenario := &alertTransitionScenario{}
	alertTransitionScenarios.Store(id, scenario)
	t.Cleanup(func() { alertTransitionScenarios.Delete(id) })
	return scenario, id
}

func (s *alertTransitionScenario) snapshot() alertTransitionSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return alertTransitionSnapshot{
		initializerCalls: s.initializerCalls,
		validateCalls:    s.validateCalls,
		resolveCalls:     s.resolveCalls,
		nextCalls:        s.nextCalls,
		adminCalls:       s.adminCalls,
		notifierCalls:    s.notifierCalls,
	}
}

func (s *alertTransitionScenario) setResolveOutcome(err error, rows []map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resolveErr = err
	s.rows = rows
}

func loadAlertTransitionScenario(id string) (*alertTransitionScenario, error) {
	value, ok := alertTransitionScenarios.Load(id)
	if !ok {
		return nil, fmt.Errorf("unknown alert transition scenario %q", id)
	}
	return value.(*alertTransitionScenario), nil
}

type alertTransitionResolver struct {
	scenario *alertTransitionScenario
}

func newAlertTransitionResolver(_ context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	id, _ := opts.Properties["scenario"].(string)
	scenario, err := loadAlertTransitionScenario(id)
	if err != nil {
		return nil, err
	}
	scenario.mu.Lock()
	defer scenario.mu.Unlock()
	scenario.initializerCalls++
	if scenario.initializerErr != nil {
		return nil, scenario.initializerErr
	}
	return &alertTransitionResolver{scenario: scenario}, nil
}

func (r *alertTransitionResolver) Close() error { return nil }

func (r *alertTransitionResolver) CacheKey(context.Context) ([]byte, bool, error) {
	return nil, false, nil
}

func (r *alertTransitionResolver) Refs() []*runtimev1.ResourceName { return nil }

func (r *alertTransitionResolver) Validate(context.Context) error {
	r.scenario.mu.Lock()
	defer r.scenario.mu.Unlock()
	r.scenario.validateCalls++
	return r.scenario.validateErr
}

func (r *alertTransitionResolver) ResolveInteractive(context.Context) (runtime.ResolverResult, error) {
	r.scenario.mu.Lock()
	defer r.scenario.mu.Unlock()
	r.scenario.resolveCalls++
	if r.scenario.resolveErr != nil {
		return nil, r.scenario.resolveErr
	}
	rows := append([]map[string]any(nil), r.scenario.rows...)
	return &alertTransitionResult{scenario: r.scenario, rows: rows, nextErr: r.scenario.nextErr}, nil
}

func (r *alertTransitionResolver) ResolveExport(context.Context, io.Writer, *runtime.ResolverExportOptions) error {
	return drivers.ErrNotImplemented
}

func (r *alertTransitionResolver) InferRequiredSecurityRules() ([]*runtimev1.SecurityRule, error) {
	return nil, nil
}

type alertTransitionResult struct {
	scenario *alertTransitionScenario
	rows     []map[string]any
	nextErr  error
	index    int
}

func (r *alertTransitionResult) Close() error                  { return nil }
func (r *alertTransitionResult) Meta() map[string]any          { return nil }
func (r *alertTransitionResult) Schema() *runtimev1.StructType { return nil }

func (r *alertTransitionResult) Next() (map[string]any, error) {
	r.scenario.mu.Lock()
	r.scenario.nextCalls++
	r.scenario.mu.Unlock()
	if r.nextErr != nil {
		return nil, r.nextErr
	}
	if r.index >= len(r.rows) {
		return nil, io.EOF
	}
	row := r.rows[r.index]
	r.index++
	return row, nil
}

func (r *alertTransitionResult) MarshalJSON() ([]byte, error) { return nil, drivers.ErrNotImplemented }

type alertTransitionDriver struct{}

func (d *alertTransitionDriver) Spec() drivers.Spec {
	return drivers.Spec{ImplementsAdmin: true, ImplementsNotifier: true}
}

func (d *alertTransitionDriver) Open(_ string, _ string, config map[string]any, _ *storage.Client, _ *activity.Client, _ *zap.Logger) (drivers.Handle, error) {
	id, _ := config["scenario"].(string)
	scenario, err := loadAlertTransitionScenario(id)
	if err != nil {
		return nil, err
	}
	return &alertTransitionHandle{scenario: scenario, config: config}, nil
}

func (d *alertTransitionDriver) HasAnonymousSourceAccess(context.Context, map[string]any, *zap.Logger) (bool, error) {
	return true, nil
}

func (d *alertTransitionDriver) TertiarySourceConnectors(context.Context, map[string]any, *zap.Logger) ([]string, error) {
	return nil, nil
}

type alertTransitionHandle struct {
	drivers.Handle
	scenario *alertTransitionScenario
	config   map[string]any
}

func (h *alertTransitionHandle) Ping(context.Context) error    { return nil }
func (h *alertTransitionHandle) Driver() string                { return alertTransitionDriverName }
func (h *alertTransitionHandle) Config() map[string]any        { return h.config }
func (h *alertTransitionHandle) Migrate(context.Context) error { return nil }
func (h *alertTransitionHandle) MigrationStatus(context.Context) (int, int, error) {
	return 0, 0, nil
}
func (h *alertTransitionHandle) Close() error { return nil }

func (h *alertTransitionHandle) AsAdmin(string) (drivers.AdminService, bool) {
	return &alertTransitionAdmin{scenario: h.scenario}, true
}

func (h *alertTransitionHandle) AsNotifier(map[string]any) (drivers.Notifier, error) {
	return &alertTransitionNotifier{scenario: h.scenario}, nil
}

type alertTransitionAdmin struct {
	drivers.AdminService
	scenario *alertTransitionScenario
}

func (a *alertTransitionAdmin) GetAlertMetadata(context.Context, string, string, []string, bool, map[string]string, string, string) (*drivers.AlertMetadata, error) {
	a.scenario.mu.Lock()
	defer a.scenario.mu.Unlock()
	a.scenario.adminCalls++
	return a.scenario.adminMeta, a.scenario.adminErr
}

func (a *alertTransitionAdmin) GetConfig(context.Context) (*drivers.Config, error) {
	return &drivers.Config{}, nil
}

type alertTransitionNotifier struct {
	scenario *alertTransitionScenario
}

func (n *alertTransitionNotifier) SendAlertStatus(*drivers.AlertStatus) error {
	n.scenario.mu.Lock()
	defer n.scenario.mu.Unlock()
	index := n.scenario.notifierCalls
	n.scenario.notifierCalls++
	if index < len(n.scenario.notifierErrs) {
		return n.scenario.notifierErrs[index]
	}
	return nil
}

func (n *alertTransitionNotifier) SendScheduledReport(*drivers.ScheduledReport) error {
	return drivers.ErrNotImplemented
}

type alertTransitionEmailSender struct {
	mu         sync.Mutex
	errs       []error
	recipients []string
}

func (s *alertTransitionEmailSender) Send(toEmail, _, _, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	index := len(s.recipients)
	s.recipients = append(s.recipients, toEmail)
	if index < len(s.errs) {
		return s.errs[index]
	}
	return nil
}

func (s *alertTransitionEmailSender) snapshot() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.recipients...)
}
