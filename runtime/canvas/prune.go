package canvas

import (
	"regexp"
	"strings"

	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/pathutil"
)

// PruneRendererProperties removes references to fields the viewer can't query from a component's renderer properties.
// canAccess reports whether a field of the component's metrics view is accessible to the viewer.
// The primary time dimension (and its encoded time-grain variants) is always allowed, since queries never check access to it.
//
// The props map is modified in place; callers should pass a copy (structpb.Struct.AsMap returns a fresh map).
// It returns false for keep when the component can't render without the restricted fields (e.g. a chart whose x or y field is restricted),
// and true for changed when any property was removed.
func PruneRendererProperties(renderer string, props map[string]any, timeDimension string, canAccess func(field string) bool) (keep, changed bool) {
	p := &fieldPruner{
		fieldRestrictions: fieldRestrictions{timeDimension: timeDimension, canAccess: canAccess},
		props:             props,
	}
	p.resolveAdhocMeasures()

	// The component's own filters apply to all of its queries, so it can't render if they reference a restricted field.
	// Dropping the restricted conditions instead would change what the component shows.
	if filter, _ := props["dimension_filters"].(string); FilterReferencesRestrictedFields(filter, timeDimension, canAccess) {
		return false, true
	}

	switch renderer {
	case "line_chart", "bar_chart", "area_chart", "stacked_bar", "stacked_bar_normalized":
		if !p.allowedAt("x.field") || !p.allowedAt("y.field") {
			return false, true
		}
		p.pruneFieldList("x.fields")
		// A multi-measure chart falls back to y.field when y.fields is empty, so it can't render if neither is left.
		if before, after := p.pruneFieldList("y.fields"); before > 0 && after == 0 {
			if _, ok := pathutil.GetPath(props, "y.field"); !ok {
				return false, true
			}
		}
		p.dropColorIfRestricted()
	case "donut_chart", "pie_chart":
		// The color field is the category dimension; the chart is meaningless without it.
		if !p.allowedAt("measure.field") || !p.allowedAt("color.field") {
			return false, true
		}
	case "scatter_plot":
		if !p.allowedAt("x.field") || !p.allowedAt("y.field") || !p.allowedAt("dimension.field") {
			return false, true
		}
		p.dropKeyIfRestricted("size.field", "size")
		p.dropColorIfRestricted()
	case "funnel_chart":
		if !p.allowedAt("stage.field") || !p.allowedAt("measure.field") {
			return false, true
		}
		if before, after := p.pruneFieldList("measure.fields"); before > 0 && after == 0 {
			if _, ok := pathutil.GetPath(props, "measure.field"); !ok {
				return false, true
			}
		}
	case "heatmap":
		if !p.allowedAt("x.field") || !p.allowedAt("y.field") || !p.allowedAt("color.field") {
			return false, true
		}
	case "combo_chart":
		if !p.allowedAt("x.field") || !p.allowedAt("y1.field") || !p.allowedAt("y2.field") {
			return false, true
		}
		p.dropColorIfRestricted()
	case "kpi":
		if !p.allowedAt("measure") {
			return false, true
		}
	case "kpi_grid":
		if before, after := p.pruneFieldList("measures"); before > 0 && after == 0 {
			return false, true
		}
	case "leaderboard":
		mBefore, mAfter := p.pruneFieldList("measures")
		dBefore, dAfter := p.pruneFieldList("dimensions")
		if (mBefore > 0 && mAfter == 0) || (dBefore > 0 && dAfter == 0) {
			return false, true
		}
	case "table":
		if before, after := p.pruneFieldList("columns"); before > 0 && after == 0 {
			return false, true
		}
		p.clearSortIfRestricted()
	case "pivot":
		mBefore, mAfter := p.pruneFieldList("measures")
		rBefore, rAfter := p.pruneFieldList("row_dimensions")
		cBefore, cAfter := p.pruneFieldList("col_dimensions")
		if mBefore+rBefore+cBefore > 0 && mAfter+rAfter+cAfter == 0 {
			return false, true
		}
		p.clearSortIfRestricted()
	case "map":
		if !p.allowedAt("geo_dimension.field") || !p.allowedAt("color.measure") {
			return false, true
		}
		p.dropKeyIfRestricted("size_measure.field", "size_measure")
		p.dropKeyIfRestricted("tooltip_dimension.field", "tooltip_dimension")
	}

	p.pruneAdhocMeasureDefinitions()
	return true, p.changed
}

// QueryReferencesRestrictedFields reports whether a metrics query selects or filters on a field that canAccess rejects.
// The query engine fails such a query, so a component that runs it can't render for the viewer.
// The primary time dimension is always allowed, matching the query engine.
func QueryReferencesRestrictedFields(q *metricsview.Query, timeDimension string, canAccess func(field string) bool) bool {
	r := fieldRestrictions{timeDimension: timeDimension, canAccess: canAccess}
	for _, d := range q.Dimensions {
		if r.dimensionRestricted(d) {
			return true
		}
	}
	for _, m := range q.Measures {
		if r.measureRestricted(m) {
			return true
		}
	}
	return r.expressionRestricted(q.Where) || r.expressionRestricted(q.Having)
}

// filterTokens matches the words and quoted strings of a component filter; punctuation and whitespace are skipped.
// Quoted strings are matched whole, including backslash escapes, so the words in a value aren't read as fields.
var filterTokens = regexp.MustCompile(`'(?:\\.|[^'\\])*'|"(?:\\.|[^"\\])*"|\w+`)

// filterOperators are the keywords that follow a field in a component filter, including the first word of NOT IN and NOT LIKE.
var filterOperators = map[string]bool{
	"IN": true, "NIN": true, "NOT": true, "HAVING": true, "LIKE": true, "NLIKE": true,
	"EQ": true, "NEQ": true, "GT": true, "GTE": true, "LT": true, "LTE": true,
}

// FilterReferencesRestrictedFields reports whether a component's own filter (its dimension_filters) references a field that canAccess rejects.
// The frontend adds the filter to each of the component's queries, so the query engine fails them if it does.
// The primary time dimension is always allowed, matching the query engine.
//
// The filter uses the syntax of dashboard URL filters, such as "country IN ('US','CA') AND city HAVING (revenue GT 100)".
// Instead of parsing that grammar (see web-common/src/features/dashboards/url-state/filters/expression.ne),
// it relies on a field always coming right before an operator, while a value never does.
// It assumes that no field is named like an operator, and that quoted field names have no escapes:
// a field named like an operator is missed, and an escaped name doesn't match its field, so it counts as restricted.
// A filter the frontend can't parse isn't applied, but the fields before its operators still count;
// that only removes a component whose filter is broken anyway.
func FilterReferencesRestrictedFields(filter, timeDimension string, canAccess func(field string) bool) bool {
	r := fieldRestrictions{timeDimension: timeDimension, canAccess: canAccess}
	tokens := filterTokens.FindAllString(filter, -1)
	for i := 0; i+1 < len(tokens); i++ {
		token := strings.ToUpper(tokens[i])
		if filterOperators[strings.ToUpper(tokens[i+1])] && !filterOperators[token] && token != "AND" && token != "OR" &&
			r.restricted(strings.Trim(tokens[i], `'"`)) {
			return true
		}
	}
	return false
}

// fieldRestrictions finds the fields of a metrics query that canAccess rejects, matching the query engine's access checks.
// The primary time dimension is always allowed.
type fieldRestrictions struct {
	timeDimension string
	canAccess     func(field string) bool
}

// restricted reports whether canAccess rejects the field.
// The primary time dimension and its time grains (e.g. "order_date_rill_TIME_GRAIN_MONTH") are always allowed, since queries never check access to them.
func (r fieldRestrictions) restricted(field string) bool {
	if field == "" || r.timeDimension != "" && (field == r.timeDimension || strings.HasPrefix(field, r.timeDimension+"_rill_")) {
		return false
	}
	return !r.canAccess(field)
}

func (r fieldRestrictions) dimensionRestricted(d metricsview.Dimension) bool {
	if d.Compute != nil && d.Compute.TimeFloor != nil {
		return r.restricted(d.Compute.TimeFloor.Dimension)
	}
	return r.restricted(d.Name)
}

func (r fieldRestrictions) measureRestricted(m metricsview.Measure) bool {
	switch {
	case m.Compute == nil:
		return r.restricted(m.Name)
	case m.Compute.CountDistinct != nil:
		return r.restricted(m.Compute.CountDistinct.Dimension)
	case m.Compute.URI != nil:
		return r.restricted(m.Compute.URI.Dimension)
	default:
		return false
	}
}

func (r fieldRestrictions) expressionRestricted(e *metricsview.Expression) bool {
	if e == nil {
		return false
	}
	if r.restricted(e.Name) {
		return true
	}
	if e.Condition != nil {
		for _, sub := range e.Condition.Expressions {
			if r.expressionRestricted(sub) {
				return true
			}
		}
	}
	if sq := e.Subquery; sq != nil {
		if r.dimensionRestricted(sq.Dimension) || r.expressionRestricted(sq.Where) || r.expressionRestricted(sq.Having) {
			return true
		}
		for _, m := range sq.Measures {
			if r.measureRestricted(m) {
				return true
			}
		}
	}
	return false
}

// fieldPruner removes restricted field references from a component's renderer properties.
type fieldPruner struct {
	fieldRestrictions
	props map[string]any
	// adhoc maps the component's adhoc (ephemeral) measure names to whether all the measures they reference are accessible.
	adhoc   map[string]bool
	changed bool
}

// allowed reports whether a field referenced by the component may be shown to the viewer.
func (p *fieldPruner) allowed(field string) bool {
	if allowed, ok := p.adhoc[field]; ok {
		return allowed
	}
	return !p.restricted(field)
}

// allowedAt reports whether the field at the path is allowed. An absent or non-string value is allowed.
func (p *fieldPruner) allowedAt(path string) bool {
	field, ok := pathutil.GetPathString(p.props, path)
	if !ok {
		return true
	}
	return p.allowed(field)
}

// pruneFieldList removes restricted fields from the list at the path, whose entries are field names or objects with a "name".
// It returns the number of entries before and after pruning.
func (p *fieldPruner) pruneFieldList(path string) (before, after int) {
	raw, ok := pathutil.GetPath(p.props, path)
	if !ok {
		return 0, 0
	}
	arr, ok := raw.([]any)
	if !ok {
		return 0, 0
	}
	kept := make([]any, 0, len(arr))
	for _, v := range arr {
		var name string
		switch t := v.(type) {
		case string:
			name = t
		case map[string]any:
			name, _ = t["name"].(string)
		}
		if name != "" && !p.allowed(name) {
			continue
		}
		kept = append(kept, v)
	}
	if len(kept) != len(arr) {
		setPath(p.props, path, kept)
		p.changed = true
	}
	return len(arr), len(kept)
}

// dropKeyIfRestricted removes the top-level property key when the field at the path is restricted.
// It's used for optional encodings, such as a scatter plot's size, which the chart can render without.
func (p *fieldPruner) dropKeyIfRestricted(path, key string) {
	if !p.allowedAt(path) {
		delete(p.props, key)
		p.changed = true
	}
}

// dropColorIfRestricted removes a color encoding that references a restricted dimension.
// Color literals (plain strings) and virtual multi-measure colors ({type: value}) don't reference fields.
func (p *fieldPruner) dropColorIfRestricted() {
	if colorType, _ := pathutil.GetPathString(p.props, "color.type"); colorType == "value" {
		return
	}
	p.dropKeyIfRestricted("color.field", "color")
}

// clearSortIfRestricted removes the initial sort of a table or pivot when it sorts by a restricted field.
func (p *fieldPruner) clearSortIfRestricted() {
	if p.allowedAt("sort_by") {
		return
	}
	delete(p.props, "sort_by")
	delete(p.props, "sort_dir")
	delete(p.props, "sort_comparison")
	p.changed = true
}

// resolveAdhocMeasures determines which of the component's adhoc measures only reference accessible measures.
func (p *fieldPruner) resolveAdhocMeasures() {
	list, _ := p.props["adhoc_measures"].([]any)
	for _, item := range list {
		entry, _ := item.(map[string]any)
		name, _ := entry["name"].(string)
		expression, _ := entry["expression"].(string)
		if name == "" {
			continue
		}
		if p.adhoc == nil {
			p.adhoc = make(map[string]bool)
		}
		parsed, err := metricsview.ParseMeasureExpression(expression)
		if err != nil {
			p.adhoc[name] = false
			continue
		}
		allowed := true
		for _, ref := range parsed.Refs() {
			if p.restricted(ref) {
				allowed = false
				break
			}
		}
		p.adhoc[name] = allowed
	}
}

// pruneAdhocMeasureDefinitions removes the definitions of adhoc measures that reference restricted measures.
func (p *fieldPruner) pruneAdhocMeasureDefinitions() {
	list, ok := p.props["adhoc_measures"].([]any)
	if !ok {
		return
	}
	kept := make([]any, 0, len(list))
	for _, item := range list {
		entry, _ := item.(map[string]any)
		name, _ := entry["name"].(string)
		if allowed, ok := p.adhoc[name]; ok && !allowed {
			continue
		}
		kept = append(kept, item)
	}
	if len(kept) != len(list) {
		p.props["adhoc_measures"] = kept
		p.changed = true
	}
}

// setPath sets a value at a dot-separated path whose parent maps already exist.
func setPath(m map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	for _, part := range parts[:len(parts)-1] {
		next, ok := m[part].(map[string]any)
		if !ok {
			return
		}
		m = next
	}
	m[parts[len(parts)-1]] = value
}
