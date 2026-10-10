package canvas

import (
	"fmt"
	"sort"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/pathutil"
	"google.golang.org/protobuf/proto"
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
		props:         props,
		timeDimension: timeDimension,
		canAccess:     canAccess,
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
		p.pruneStringList("x.fields")
		p.pruneStringList("y.fields")
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
		if before, after := p.pruneStringList("measure.fields"); before > 0 && after == 0 {
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
		if before, after := p.pruneStringList("measures"); before > 0 && after == 0 {
			return false, true
		}
	case "leaderboard":
		mBefore, mAfter := p.pruneStringList("measures")
		dBefore, dAfter := p.pruneStringList("dimensions")
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

// FilterReferencesRestrictedFields reports whether a component's own filter (its dimension_filters) references a field that canAccess rejects.
// The frontend adds the filter to each of the component's queries, so the query engine fails them if it does.
// The frontend doesn't apply a filter that it can't parse, so such a filter doesn't reference any fields.
// The primary time dimension is always allowed, matching the query engine.
func FilterReferencesRestrictedFields(filter, timeDimension string, canAccess func(field string) bool) bool {
	expr, err := parseFilter(filter)
	if err != nil {
		return false
	}
	r := fieldRestrictions{timeDimension: timeDimension, canAccess: canAccess}
	return r.expressionRestricted(expr)
}

// fieldRestrictions finds the fields of a metrics query that canAccess rejects, matching the query engine's access checks.
// The primary time dimension is always allowed.
type fieldRestrictions struct {
	timeDimension string
	canAccess     func(field string) bool
}

func (r fieldRestrictions) restricted(field string) bool {
	return field != "" && field != r.timeDimension && !r.canAccess(field)
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

// Keep decides which canvas elements PruneRows keeps. A nil function keeps every element of its kind.
type Keep struct {
	// Row is called for plain rows and for tab-group rows.
	Row  func(row *runtimev1.CanvasRow) bool
	Tab  func(tab *runtimev1.CanvasTab) bool
	Item func(item *runtimev1.CanvasItem) bool
}

// PruneRows returns the rows without the elements that keep rejects,
// and the YAML paths of the rejected elements, such as "rows.1", "rows.0.items.1" or "rows.3.tabs.0.rows.0.items.0".
// The remaining items in a row keep their relative widths, rescaled to the row's original total.
// Rows, tabs and tab groups that lose all of their content are removed too; ones that were authored empty are kept.
// The input rows are not modified; if nothing is rejected, they are returned as is.
func PruneRows(rows []*runtimev1.CanvasRow, keep Keep) ([]*runtimev1.CanvasRow, []string) {
	var paths []string
	res := pruneRows(rows, keep, "rows", &paths)
	return res, paths
}

// pruneRows implements PruneRows for rows at the given YAML path prefix, appending the paths of rejected elements to paths.
func pruneRows(rows []*runtimev1.CanvasRow, keep Keep, prefix string, paths *[]string) []*runtimev1.CanvasRow {
	initial := len(*paths)
	var out []*runtimev1.CanvasRow
	for i, row := range rows {
		rowPath := fmt.Sprintf("%s.%d", prefix, i)
		if keep.Row != nil && !keep.Row(row) {
			*paths = append(*paths, rowPath)
			continue
		}

		if tg := row.GetTabGroup(); tg != nil {
			groupInitial := len(*paths)
			var tabs []*runtimev1.CanvasTab
			for t, tab := range tg.Tabs {
				tabPath := fmt.Sprintf("%s.tabs.%d", rowPath, t)
				if keep.Tab != nil && !keep.Tab(tab) {
					*paths = append(*paths, tabPath)
					continue
				}
				tabInitial := len(*paths)
				tabRows := pruneRows(tab.Rows, keep, tabPath+".rows", paths)
				if len(*paths) == tabInitial {
					tabs = append(tabs, tab)
					continue
				}
				if len(tabRows) == 0 {
					continue
				}
				tab = proto.Clone(tab).(*runtimev1.CanvasTab)
				tab.Rows = tabRows
				tabs = append(tabs, tab)
			}
			if len(*paths) == groupInitial {
				out = append(out, row)
				continue
			}
			if len(tabs) == 0 {
				continue
			}
			row = proto.Clone(row).(*runtimev1.CanvasRow)
			row.TabGroup.Tabs = tabs
			out = append(out, row)
			continue
		}

		var items []*runtimev1.CanvasItem
		for j, item := range row.Items {
			if keep.Item != nil && !keep.Item(item) {
				*paths = append(*paths, fmt.Sprintf("%s.items.%d", rowPath, j))
				continue
			}
			items = append(items, item)
		}
		if len(items) == len(row.Items) {
			out = append(out, row)
			continue
		}
		if len(items) == 0 {
			continue
		}
		newRow := proto.Clone(row).(*runtimev1.CanvasRow)
		newRow.Items = rescaleItemWidths(row.Items, items)
		out = append(out, newRow)
	}
	if len(*paths) == initial {
		return rows
	}
	return out
}

// StripConditions clears the `if` conditions of the rows, tabs and items, so they aren't sent to viewers.
// It modifies the rows in place.
func StripConditions(rows []*runtimev1.CanvasRow) {
	for _, row := range rows {
		row.ConditionExpression = ""
		for _, item := range row.Items {
			item.ConditionExpression = ""
		}
		if tg := row.GetTabGroup(); tg != nil {
			for _, tab := range tg.Tabs {
				tab.ConditionExpression = ""
				StripConditions(tab.Rows)
			}
		}
	}
}

// Conditions returns the distinct `if` conditions of the rows, tabs and items.
func Conditions(rows []*runtimev1.CanvasRow) []string {
	seen := make(map[string]bool)
	var res []string
	add := func(condition string) {
		if condition != "" && !seen[condition] {
			seen[condition] = true
			res = append(res, condition)
		}
	}
	var walk func(rows []*runtimev1.CanvasRow)
	walk = func(rows []*runtimev1.CanvasRow) {
		for _, row := range rows {
			add(row.ConditionExpression)
			for _, item := range row.Items {
				add(item.ConditionExpression)
			}
			if tg := row.GetTabGroup(); tg != nil {
				for _, tab := range tg.Tabs {
					add(tab.ConditionExpression)
					walk(tab.Rows)
				}
			}
		}
	}
	walk(rows)
	return res
}

// rescaleItemWidths rescales the widths of the kept items so they add up to the total width of all the row's items, preserving their proportions.
// It uses the largest remainder method so the widths stay whole numbers.
// If any item has no width (the frontend then splits the row equally), the kept items are returned unchanged.
func rescaleItemWidths(all, kept []*runtimev1.CanvasItem) []*runtimev1.CanvasItem {
	var total, keptTotal uint32
	for _, item := range all {
		if item.Width == nil || *item.Width == 0 || item.WidthUnit != "" {
			return kept
		}
		total += *item.Width
	}
	for _, item := range kept {
		keptTotal += *item.Width
	}
	if keptTotal == 0 || keptTotal == total {
		return kept
	}

	type share struct {
		idx       int
		width     uint32
		remainder uint32
	}
	shares := make([]share, len(kept))
	var assigned uint32
	for i, item := range kept {
		n := *item.Width * total
		shares[i] = share{idx: i, width: n / keptTotal, remainder: n % keptTotal}
		assigned += shares[i].width
	}
	byRemainder := make([]share, len(shares))
	copy(byRemainder, shares)
	sort.SliceStable(byRemainder, func(i, j int) bool { return byRemainder[i].remainder > byRemainder[j].remainder })
	for i := 0; assigned < total; i++ {
		shares[byRemainder[i%len(byRemainder)].idx].width++
		assigned++
	}

	res := make([]*runtimev1.CanvasItem, len(kept))
	for i, item := range kept {
		item = proto.Clone(item).(*runtimev1.CanvasItem)
		w := shares[i].width
		item.Width = &w
		res[i] = item
	}
	return res
}

// fieldPruner removes restricted field references from a component's renderer properties.
type fieldPruner struct {
	props         map[string]any
	timeDimension string
	canAccess     func(field string) bool
	// adhoc maps the component's adhoc (ephemeral) measure names to whether all the measures they reference are accessible.
	adhoc   map[string]bool
	changed bool
}

// allowed reports whether a field referenced by the component may be shown to the viewer.
func (p *fieldPruner) allowed(field string) bool {
	if allowed, ok := p.adhoc[field]; ok {
		return allowed
	}
	if p.timeDimension != "" && (field == p.timeDimension || strings.HasPrefix(field, p.timeDimension+"_rill_")) {
		return true
	}
	return p.canAccess(field)
}

// allowedAt reports whether the field at the path is allowed. An absent or non-string value is allowed.
func (p *fieldPruner) allowedAt(path string) bool {
	field, ok := pathutil.GetPathString(p.props, path)
	if !ok {
		return true
	}
	return p.allowed(field)
}

// pruneStringList removes restricted fields from the array of field names at the path.
// It returns the number of entries before and after pruning.
func (p *fieldPruner) pruneStringList(path string) (before, after int) {
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
		if s, ok := v.(string); ok && !p.allowed(s) {
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

// pruneFieldList removes restricted fields from a table or pivot field list, whose entries are field names or objects with a "name".
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
	raw, ok := p.props["color"]
	if !ok {
		return
	}
	if _, isString := raw.(string); isString {
		return
	}
	if colorType, ok := pathutil.GetPathString(p.props, "color.type"); ok && colorType == "value" {
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
			if !p.canAccess(ref) {
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
