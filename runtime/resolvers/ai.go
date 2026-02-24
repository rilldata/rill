package resolvers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
)

func init() {
	runtime.RegisterResolverInitializer("ai", newAI)
}

// aiProps contains the static properties for the AI resolver.
type aiProps struct {
	Agent               string                 `mapstructure:"agent"`
	Prompt              string                 `mapstructure:"prompt"`
	TimeRange           *metricsview.TimeRange `mapstructure:"time_range"` // Time range for analysis (supports rilltime expressions, or fixed start/end)
	ComparisonTimeRange *metricsview.TimeRange `mapstructure:"comparison_time_range"`
	TimeZone            string                 `mapstructure:"time_zone"`
	// optional dashboard context for the agent
	Explore    string         `mapstructure:"explore"`
	Dimensions []string       `mapstructure:"dimensions"`
	Measures   []string       `mapstructure:"measures"`
	Where      map[string]any `mapstructure:"where"`
	// IsReport indicates if the AI resolver is used for an automated report.
	IsReport bool `mapstructure:"is_report"`
}

// aiArgs contains the dynamic arguments for the AI resolver.
type aiArgs struct {
	// ExecutionTime used to resolve time ranges
	ExecutionTime time.Time `mapstructure:"execution_time"`
	// CreateSharedSession indicates if a shared session should be created.
	CreateSharedSession bool `mapstructure:"create_shared_session"`
}

// newAI creates a new AI resolver.
func newAI(ctx context.Context, opts *runtime.ResolverOptions) (runtime.Resolver, error) {
	// Parse props
	props := &aiProps{}
	if err := mapstructure.Decode(opts.Properties, props); err != nil {
		return nil, err
	}

	// Parse args
	args := &aiArgs{}
	if err := mapstructure.Decode(opts.Args, args); err != nil {
		return nil, err
	}

	// Default execution time to now
	if args.ExecutionTime.IsZero() {
		args.ExecutionTime = time.Now()
	}

	if props.Agent != ai.AnalystAgentName {
		return nil, errors.New("only 'analyst_agent' is supported as agent as of now")
	}

	if !props.IsReport && props.Prompt == "" {
		return nil, errors.New("prompt is required for non-report AI sessions")
	}

	if props.TimeRange != nil && (props.TimeRange.IsoDuration != "" || props.TimeRange.IsoOffset != "") {
		return nil, errors.New("iso_duration and iso_offset are deprecated in favor of rilltime expressions")
	}

	if props.ComparisonTimeRange != nil && (props.ComparisonTimeRange.IsoDuration != "" || props.ComparisonTimeRange.IsoOffset != "") {
		return nil, errors.New("iso_duration and iso_offset are deprecated in favor of rilltime expressions")
	}

	// Get metrics view if explore is provided
	var mv string
	if props.Explore != "" {
		c, err := opts.Runtime.Controller(ctx, opts.InstanceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get controller: %w", err)
		}
		e, err := c.Get(ctx, &runtimev1.ResourceName{
			Kind: runtime.ResourceKindExplore,
			Name: props.Explore,
		}, false)
		if err != nil {
			return nil, fmt.Errorf("failed to get explore %q: %w", props.Explore, err)
		}
		exp := e.GetExplore()
		if exp == nil {
			return nil, fmt.Errorf("resource %q is not an explore", props.Explore)
		}
		spec := exp.State.ValidSpec
		if spec == nil {
			return nil, fmt.Errorf("explore %q has no valid spec", props.Explore)
		}
		mv = spec.MetricsView
	}

	return &aiResolver{
		runtime:     opts.Runtime,
		instanceID:  opts.InstanceID,
		props:       props,
		args:        args,
		claims:      opts.Claims,
		metricsView: mv,
	}, nil
}

type aiResolver struct {
	runtime     *runtime.Runtime
	instanceID  string
	props       *aiProps
	args        *aiArgs
	claims      *runtime.SecurityClaims
	metricsView string // optional metrics view spec if available
}

var _ runtime.Resolver = &aiResolver{}

// Close implements runtime.Resolver.
func (r *aiResolver) Close() error {
	return nil
}

// CacheKey implements runtime.Resolver.
// AI sessions are not cacheable since they produce unique sessions each time.
func (r *aiResolver) CacheKey(ctx context.Context) ([]byte, bool, error) {
	return nil, false, nil
}

// Refs implements runtime.Resolver.
func (r *aiResolver) Refs() []*runtimev1.ResourceName {
	return nil
}

// Validate implements runtime.Resolver.
func (r *aiResolver) Validate(ctx context.Context) error {
	if r.props.Agent != ai.AnalystAgentName {
		return errors.New("only 'analyst_agent' is supported as agent as of now")
	}
	if !r.props.IsReport && r.props.Prompt == "" {
		return errors.New("prompt is required for non-report AI sessions")
	}
	return nil
}

// ResolveInteractive implements runtime.Resolver.
func (r *aiResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	// Resolve time ranges if provided
	err := r.resolveTimeRange(ctx, r.props.TimeRange, r.props.TimeZone)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve time range: %w", err)
	}

	// Resolve comparison time range if provided
	err = r.resolveTimeRange(ctx, r.props.ComparisonTimeRange, r.props.TimeZone)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve comparison time range: %w", err)
	}

	var whereExpr *metricsview.Expression
	explore := r.props.Explore
	dimensions := r.props.Dimensions
	measures := r.props.Measures
	if len(r.props.Where) > 0 {
		whereExpr = &metricsview.Expression{}
		if err := mapstructure.Decode(r.props.Where, whereExpr); err != nil {
			return nil, fmt.Errorf("failed to parse where filter: %w", err)
		}
	}

	runner := ai.NewRunner(r.runtime, r.runtime.Activity())

	// Create a new AI session
	session, err := runner.Session(ctx, &ai.SessionOptions{
		InstanceID:        r.instanceID,
		CreateIfNotExists: true,
		Claims:            r.claims,
		UserAgent:         "rill/report", //  rill/report agent will be filtered in the UI conversation listing
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AI session: %w", err)
	}
	defer session.Flush(ctx)

	agentArgs := &ai.AnalystAgentArgs{
		Explore:             explore,
		Dimensions:          dimensions,
		Measures:            measures,
		Where:               whereExpr,
		TimeStart:           r.props.TimeRange.Start,
		TimeEnd:             r.props.TimeRange.End,
		ComparisonTimeStart: r.props.ComparisonTimeRange.Start,
		ComparisonTimeEnd:   r.props.ComparisonTimeRange.End,
		DisableCharts:       r.args.CreateSharedSession, // Disable charts if creating shared session
		IsReport:            r.props.IsReport,
	}

	routerArgs := &ai.RouterAgentArgs{
		Prompt:           r.props.Prompt,
		Agent:            r.props.Agent,
		AnalystAgentArgs: agentArgs,
	}

	// Call the analyst agent
	var result ai.RouterAgentResult
	_, err = session.CallTool(ctx, ai.RoleUser, "router_agent", &result, routerArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to call agent: %w", err)
	}

	// Update session title
	err = session.UpdateTitle(ctx, r.generateTitle())
	if err != nil {
		return nil, fmt.Errorf("failed to update session title: %w", err)
	}

	if r.args.CreateSharedSession {
		msg, ok := session.LatestMessage([]ai.Predicate{ai.FilterByTool(ai.RouterAgentName), ai.FilterByType(ai.MessageTypeResult)}...)
		if !ok {
			return nil, fmt.Errorf("failed to create shared session: no result message found")
		}
		err = session.UpdateSharedUntilMessageID(ctx, msg.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to create shared session: %w", err)
		}
	}

	// Extract summary from the response (from <summary> tag or fallback to truncation)
	summary := extractSummary(result.Response)

	// Return the session ID and summary
	rows := []map[string]any{
		{
			"ai_session_id": session.ID(),
			"summary":       summary,
			"title":         session.Title(),
		},
	}

	schema := &runtimev1.StructType{
		Fields: []*runtimev1.StructType_Field{
			{Name: "ai_session_id", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
			{Name: "summary", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
			{Name: "title", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}},
		},
	}

	return runtime.NewMapsResolverResult(rows, schema), nil
}

// ResolveExport implements runtime.Resolver.
func (r *aiResolver) ResolveExport(ctx context.Context, w io.Writer, opts *runtime.ResolverExportOptions) error {
	return fmt.Errorf("AI resolver does not support export")
}

// InferRequiredSecurityRules implements runtime.Resolver.
func (r *aiResolver) InferRequiredSecurityRules() ([]*runtimev1.SecurityRule, error) {
	return nil, nil
}

// resolveTimeRange resolves and rewrites the time range to actual timestamps using rilltime.
func (r *aiResolver) resolveTimeRange(ctx context.Context, tr *metricsview.TimeRange, tz string) error {
	// if time range is not provided, or already has start and end, do nothing
	if tr == nil || tr.IsZero() || (!tr.Start.IsZero() && !tr.End.IsZero()) {
		return nil
	}

	var err error
	timezone := time.UTC
	if tz != "" {
		timezone, err = time.LoadLocation(tz)
		if err != nil {
			return fmt.Errorf("invalid time zone %q: %w", tz, err)
		}
	}

	// if metrics view is provided, we can get the metrics view's time bounds
	if r.metricsView != "" {
		mv, security, err := resolveMVAndSecurityFromAttributes(ctx, r.runtime, r.instanceID, r.metricsView, r.claims)
		if err != nil {
			return fmt.Errorf("failed to resolve metrics view %q: %w", r.metricsView, err)
		}
		// create executor to resolve relative time ranges
		e, err := executor.New(ctx, r.runtime, r.instanceID, mv.ValidSpec, mv.Streaming, security, 10, r.claims.UserAttributes)
		if err != nil {
			return fmt.Errorf("failed to create executor: %w", err)
		}
		return e.ResolveTimeRange(ctx, tr, timezone, &r.args.ExecutionTime)
	}

	// Without explore/metrics view, we can only use execution time to resolve time ranges
	if r.args.ExecutionTime.IsZero() {
		return errors.New("execution_time is required to evaluate time ranges without explore context")
	}

	// Use expression if provided (rilltime syntax)
	if tr.Expression != "" {
		rt, err := rilltime.Parse(tr.Expression, rilltime.ParseOptions{
			DefaultTimeZone: timezone,
		})
		if err != nil {
			return fmt.Errorf("invalid time range expression %q: %w", tr.Expression, err)
		}

		start, end, _ := rt.Eval(rilltime.EvalOptions{
			Now:       time.Now(),
			Watermark: r.args.ExecutionTime,
			MinTime:   time.Time{},
			MaxTime:   r.args.ExecutionTime,
		})
		tr.Start = start
		tr.End = end
		// Clear other fields
		tr.Expression = ""
		tr.IsoDuration = ""
		tr.IsoOffset = ""
		tr.RoundToGrain = metricsview.TimeGrainUnspecified
		return nil
	}

	return errors.New("time range must have either start/end or expression")
}

// generateTitle generates a title for the AI session.
func (r *aiResolver) generateTitle() string {
	title := "AI Session"
	if r.props.IsReport {
		title = "Report"
	}
	title = fmt.Sprintf("%s - %s", title, r.args.ExecutionTime.Format(time.RFC822))
	if r.props.Explore != "" {
		return fmt.Sprintf("%s: %s", title, r.props.Explore)
	}
	return title
}

func resolveMVAndSecurityFromAttributes(ctx context.Context, rt *runtime.Runtime, instanceID, metricsViewName string, claims *runtime.SecurityClaims) (*runtimev1.MetricsViewState, *runtime.ResolvedSecurity, error) {
	res, mv, err := lookupMetricsView(ctx, rt, instanceID, metricsViewName)
	if err != nil {
		return nil, nil, err
	}

	resolvedSecurity, err := rt.ResolveSecurity(ctx, instanceID, claims, res)
	if err != nil {
		return nil, nil, err
	}

	if !resolvedSecurity.CanAccess() {
		return nil, nil, runtime.ErrForbidden
	}

	return mv, resolvedSecurity, nil
}

// returns the metrics view and the time the catalog was last updated
func lookupMetricsView(ctx context.Context, rt *runtime.Runtime, instanceID, name string) (*runtimev1.Resource, *runtimev1.MetricsViewState, error) {
	ctrl, err := rt.Controller(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}

	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: name}, false)
	if err != nil {
		return nil, nil, err
	}

	mv := res.GetMetricsView()
	spec := mv.State.ValidSpec
	if spec == nil {
		return nil, nil, fmt.Errorf("metrics view %q is invalid", name)
	}

	return res, mv.State, nil
}

// extractSummary extracts the summary from the <summary> tag in the AI response.
// If no summary tag is found, it returns an empty string.
func extractSummary(response string) string {
	// Look for <summary>...</summary> pattern
	start := strings.Index(response, "<summary>")
	end := strings.Index(response, "</summary>")
	if start != -1 && end != -1 && end > start {
		return strings.TrimSpace(response[start+9 : end])
	}
	return ""
}
