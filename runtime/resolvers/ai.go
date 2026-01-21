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
)

func init() {
	runtime.RegisterResolverInitializer("ai", newAI)
}

// aiProps contains the static properties for the AI resolver.
type aiProps struct {
	Agent  string `mapstructure:"agent"`
	Prompt string `mapstructure:"prompt"`
	// Relative time range configuration may add exact start/end in the future
	TimeRangeISODuration string `mapstructure:"time_range_iso_duration"`
	TimeRangeTimeZone    string `mapstructure:"time_range_time_zone"`
	// Optional comparison time range
	ComparisonTimeRangeISODuration string `mapstructure:"comparison_time_range_iso_duration"`
	ComparisonTimeRangeISOOffset   string `mapstructure:"comparison_time_range_iso_offset"`
	// Optional dashboard context for the agent
	Explore    string         `mapstructure:"explore"`
	Dimensions []string       `mapstructure:"dimensions"`
	Measures   []string       `mapstructure:"measures"`
	Where      map[string]any `mapstructure:"where"`
	// IsScheduledInsight indicates if the AI resolver is used for a scheduled insight.
	IsScheduledInsight bool `mapstructure:"is_scheduled_insight"`
}

// aiArgs contains the dynamic arguments for the AI resolver.
type aiArgs struct {
	// ExecutionTime used to resolve time ranges
	ExecutionTime time.Time `mapstructure:"execution_time"`
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

	return &aiResolver{
		runtime:    opts.Runtime,
		instanceID: opts.InstanceID,
		props:      props,
		args:       args,
		claims:     opts.Claims,
	}, nil
}

type aiResolver struct {
	runtime    *runtime.Runtime
	instanceID string
	props      *aiProps
	args       *aiArgs
	claims     *runtime.SecurityClaims
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
	if r.props.Explore != "" {
		refs = append(refs, &runtimev1.ResourceName{
			Kind: runtime.ResourceKindExplore,
			Name: r.props.Explore,
		})
	}
	return refs
}

// Validate implements runtime.Resolver.
func (r *aiResolver) Validate(ctx context.Context) error {
	if r.props.Agent != ai.AnalystAgentName {
		return errors.New("only 'analyst_agent' is supported as agent as of now")
	}
	if r.props.Prompt == "" {
		return errors.New("prompt is required")
	}
	return nil
}

// ResolveInteractive implements runtime.Resolver.
func (r *aiResolver) ResolveInteractive(ctx context.Context) (runtime.ResolverResult, error) {
	// Resolve time ranges
	timeStart, timeEnd, err := r.resolveTimeRange()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve time range: %w", err)
	}

	// Resolve comparison time range
	comparisonStart, comparisonEnd, err := r.resolveComparisonTimeRange(timeStart)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve comparison time range: %w", err)
	}

	runner := ai.NewRunner(r.runtime, r.runtime.Activity())

	// Create a new AI session
	session, err := runner.Session(ctx, &ai.SessionOptions{
		InstanceID:        r.instanceID,
		CreateIfNotExists: true,
		Claims:            r.claims,
		UserAgent:         "rill/report", // TODO change it to system/report or similar so that its not shown in AI sessions list
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AI session: %w", err)
	}
	defer session.Flush(ctx)

	// Parse where filter if provided
	var whereExpr *metricsview.Expression
	if len(r.props.Where) > 0 {
		whereExpr = &metricsview.Expression{}
		if err := mapstructure.Decode(r.props.Where, whereExpr); err != nil {
			return nil, fmt.Errorf("failed to parse where filter: %w", err)
		}
	}

	// Build analyst agent args with all time ranges
	agentArgs := &ai.AnalystAgentArgs{
		Explore:             r.props.Explore,
		Dimensions:          r.props.Dimensions,
		Measures:            r.props.Measures,
		Where:               whereExpr,
		TimeStart:           timeStart,
		TimeEnd:             timeEnd,
		ComparisonTimeStart: comparisonStart,
		ComparisonTimeEnd:   comparisonEnd,
		IsScheduledInsight:  r.props.IsScheduledInsight,
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

// resolveTimeRange resolves the time range from ISO duration to actual timestamps.
func (r *aiResolver) resolveTimeRange() (start, end time.Time, err error) {
	// Load timezone
	loc := time.UTC
	if r.props.TimeRangeTimeZone != "" {
		loc, err = time.LoadLocation(r.props.TimeRangeTimeZone)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid timezone %q: %w", r.props.TimeRangeTimeZone, err)
		}
	}

	// Convert execution time to the specified timezone
	execInTZ := r.args.ExecutionTime.In(loc)

	// End is truncated to start of day
	end = time.Date(execInTZ.Year(), execInTZ.Month(), execInTZ.Day(), 0, 0, 0, 0, loc)

	// Parse duration and subtract from end to get start
	dur, err := duration.ParseISO8601(r.props.TimeRangeISODuration)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid ISO duration %q: %w", r.props.TimeRangeISODuration, err)
	}
	start = dur.Sub(end)

	return start, end, nil
}

// resolveComparisonTimeRange resolves the comparison time range.
func (r *aiResolver) resolveComparisonTimeRange(mainTimeStart time.Time) (start, end time.Time, err error) {
	if r.props.ComparisonTimeRangeISODuration == "" {
		return time.Time{}, time.Time{}, nil
	}
	// End of comparison = start of main time range
	end = mainTimeStart
	dur, err := duration.ParseISO8601(r.props.ComparisonTimeRangeISODuration)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid comparison duration %q: %w", r.props.ComparisonTimeRangeISODuration, err)
	}
	start = dur.Sub(end)

	// Apply offset if provided
	if r.props.ComparisonTimeRangeISOOffset != "" {
		d, err := duration.ParseISO8601(r.props.ComparisonTimeRangeISOOffset)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid comparison offset %q: %w", r.props.ComparisonTimeRangeISOOffset, err)
		}
		start = d.Sub(start)
		end = d.Sub(end)
	}

	return start, end, nil
}

// generateTitle generates a title for the AI session.
func (r *aiResolver) generateTitle() string {
	if r.props.Explore != "" {
		return fmt.Sprintf("Scheduled Insight: %s", r.props.Explore)
	}
	return "Scheduled Insight Report"
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
