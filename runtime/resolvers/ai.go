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
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

func init() {
	runtime.RegisterResolverInitializer("ai", newAI)
}

// aiProps contains the static properties for the AI resolver.
type aiProps struct {
	Agent  string `mapstructure:"agent"`
	Prompt string `mapstructure:"prompt"`
	// Time range for analysis (supports rilltime expressions, ISO durations, or fixed start/end)
	TimeRange *metricsview.TimeRange `mapstructure:"time_range"`
	// Optional comparison time range
	ComparisonTimeRange *metricsview.TimeRange `mapstructure:"comparison_time_range"`
	TimeZone            string                 `mapstructure:"time_zone"`
	// Optional dashboard context for the agent
	Context *contextualProps `mapstructure:"context"`
	// IsReport indicates if the AI resolver is used for an automated report.
	IsReport bool `mapstructure:"is_report"`
}

type contextualProps struct {
	Explore    string         `mapstructure:"explore"`
	Dimensions []string       `mapstructure:"dimensions"`
	Measures   []string       `mapstructure:"measures"`
	Where      map[string]any `mapstructure:"where"`
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

	// Get metrics view if explore is provided
	var mv string
	if props.Context != nil && props.Context.Explore != "" {
		c, err := opts.Runtime.Controller(ctx, opts.InstanceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get controller: %w", err)
		}
		e, err := c.Get(ctx, &runtimev1.ResourceName{
			Kind: runtime.ResourceKindExplore,
			Name: props.Context.Explore,
		}, false)
		if err != nil {
			return nil, fmt.Errorf("failed to get explore %q: %w", props.Context.Explore, err)
		}
		exp := e.GetExplore()
		if exp == nil {
			return nil, fmt.Errorf("resource %q is not an explore", props.Context.Explore)
		}
		spec := exp.State.ValidSpec
		if spec == nil {
			return nil, fmt.Errorf("explore %q has no valid spec", props.Context.Explore)
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
	var refs []*runtimev1.ResourceName
	if r.props.Context != nil && r.props.Context.Explore != "" {
		refs = append(refs, &runtimev1.ResourceName{
			Kind: runtime.ResourceKindExplore,
			Name: r.props.Context.Explore,
		})
	}
	if r.metricsView != "" {
		refs = append(refs, &runtimev1.ResourceName{
			Kind: runtime.ResourceKindMetricsView,
			Name: r.metricsView,
		})
	}
	return refs
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
	timeStart, timeEnd, err := r.resolveTimeRange(r.props.TimeRange, r.props.TimeZone)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve time range: %w", err)
	}

	// Resolve comparison time range if provided
	comparisonStart, comparisonEnd, err := r.resolveTimeRange(r.props.ComparisonTimeRange, r.props.TimeZone)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve comparison time range: %w", err)
	}

	var explore string
	var dimensions, measures []string
	var whereExpr *metricsview.Expression
	if r.props.Context != nil {
		explore = r.props.Context.Explore
		dimensions = r.props.Context.Dimensions
		measures = r.props.Context.Measures
		if len(r.props.Context.Where) > 0 {
			whereExpr = &metricsview.Expression{}
			if err := mapstructure.Decode(r.props.Context.Where, whereExpr); err != nil {
				return nil, fmt.Errorf("failed to parse where filter: %w", err)
			}
		}
	}

	runner := ai.NewRunner(r.runtime, r.runtime.Activity())

	// Create a new AI session
	session, err := runner.Session(ctx, &ai.SessionOptions{
		InstanceID:        r.instanceID,
		CreateIfNotExists: true,
		Claims:            r.claims,
		UserAgent:         "rill/report", // TODO change it to system/report or similar so that its not shown in AI sessions list, keeping it rill prefixed for now so that access checks pass
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
		TimeStart:           timeStart,
		TimeEnd:             timeEnd,
		ComparisonTimeStart: comparisonStart,
		ComparisonTimeEnd:   comparisonEnd,
		DisableCharts:       r.args.CreateSharedSession, // Disable charts if creating shared session
		IsReport:            r.props.IsReport,
		IsReportUserPrompt:  r.props.Prompt != "",
	}

	prompt := r.props.Prompt
	if r.props.IsReport && prompt == "" {
		prompt = "Generate the scheduled insight report."
	}

	routerArgs := &ai.RouterAgentArgs{
		Prompt:           prompt,
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
	return nil, errors.New("security rule inference not implemented yet for AI resolver")
}

// resolveTimeRange resolves the time range to actual timestamps using rilltime.
func (r *aiResolver) resolveTimeRange(tr *metricsview.TimeRange, tz string) (start, end time.Time, err error) {
	if tr == nil || tr.IsZero() {
		return time.Time{}, time.Time{}, nil
	}

	timezone := time.UTC
	if tz != "" {
		timezone, err = time.LoadLocation(tz)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid time zone %q: %w", tz, err)
		}
	}

	// If start and end are already set, return them directly
	if !tr.Start.IsZero() && !tr.End.IsZero() {
		return tr.Start, tr.End, nil
	}

	if r.args.ExecutionTime.IsZero() {
		return time.Time{}, time.Time{}, errors.New("execution_time is required to evaluate time ranges")
	}

	// Use expression if provided (rilltime syntax)
	if tr.Expression != "" {
		rt, err := rilltime.Parse(tr.Expression, rilltime.ParseOptions{
			DefaultTimeZone: timezone,
		})
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid time range expression %q: %w", tr.Expression, err)
		}

		start, end, _ = rt.Eval(rilltime.EvalOptions{
			Now:       r.args.ExecutionTime,
			Watermark: r.args.ExecutionTime,
			MinTime:   time.Time{}, // TODO if context contains explore then resolve min time from the metrics view
			MaxTime:   r.args.ExecutionTime,
		})
		return start, end, nil
	}

	// TODO how about we don't support ISO duration/offset as its deprecated in favor of rilltime expressions?

	// Fallback to start/end with ISO duration/offset
	start = tr.Start
	end = tr.End
	isISO := false
	if start.IsZero() && end.IsZero() {
		end = r.args.ExecutionTime
	}

	if tr.IsoDuration != "" {
		d, err := duration.ParseISO8601(tr.IsoDuration)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid iso_duration %q: %w", tr.IsoDuration, err)
		}

		if !start.IsZero() && !end.IsZero() {
			return time.Time{}, time.Time{}, errors.New(`cannot resolve "iso_duration" for a time range with fixed "start" and "end" timestamps`)
		} else if !start.IsZero() {
			end = d.Add(start)
		} else if !end.IsZero() {
			start = d.Sub(end)
		}
		isISO = true
	}

	// Apply offset if provided
	if tr.IsoOffset != "" {
		d, err := duration.ParseISO8601(tr.IsoOffset)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid iso_offset %q: %w", tr.IsoOffset, err)
		}

		if !start.IsZero() {
			start = d.Sub(start)
		}
		if !end.IsZero() {
			end = d.Sub(end)
		}
		isISO = true
	}

	// Only modify the start and end if ISO duration or offset was sent.
	// This is to maintain backwards compatibility for calls from the UI.
	if isISO {
		if !tr.RoundToGrain.Valid() {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid time grain %q", tr.RoundToGrain)
		}
		if tr.RoundToGrain != metricsview.TimeGrainUnspecified {
			if !start.IsZero() {
				start = timeutil.TruncateTime(start, tr.RoundToGrain.ToTimeutil(), timezone, 1, 1)
			}
			if !end.IsZero() {
				end = timeutil.TruncateTime(end, tr.RoundToGrain.ToTimeutil(), timezone, 1, 1)
			}
		}
		return start, end, nil
	}

	return time.Time{}, time.Time{}, errors.New("time range must have expression, iso_duration, or start/end")
}

// generateTitle generates a title for the AI session.
func (r *aiResolver) generateTitle() string {
	title := "AI Session"
	if r.props.IsReport {
		title = "Report"
	}
	title = fmt.Sprintf("%s - %s", title, r.args.ExecutionTime.Format(time.RFC822))
	if r.props.Context != nil && r.props.Context.Explore != "" {
		return fmt.Sprintf("%s: %s", title, r.props.Context.Explore)
	}
	return title
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
