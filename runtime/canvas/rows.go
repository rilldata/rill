package canvas

import (
	"fmt"
	"slices"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/proto"
)

// Keep decides which canvas elements PruneRows keeps. A nil function keeps every element of its kind.
type Keep struct {
	// Row is called for plain rows, tab-group rows and container rows.
	Row  func(row *runtimev1.CanvasRow) bool
	Tab  func(tab *runtimev1.CanvasTab) bool
	Item func(item *runtimev1.CanvasItem) bool
}

// PruneRows returns the rows without the elements that keep rejects,
// and the YAML paths of the rejected elements, such as "rows.1", "rows.0.items.1", "rows.3.tabs.0.rows.0.items.0" or "rows.2.rows.1".
// Rows, tabs, tab groups and containers left without content are removed too.
// The remaining items keep their widths; the frontend splits a row equally when its widths no longer add up to the full width.
// The input rows are not modified, and they're returned as is if nothing is removed.
func PruneRows(rows []*runtimev1.CanvasRow, keep Keep) ([]*runtimev1.CanvasRow, []string) {
	var paths []string
	res := pruneRows(rows, keep, "rows", &paths)
	return res, paths
}

// pruneRows implements PruneRows for rows at the given YAML path prefix, appending the paths of rejected elements to paths.
func pruneRows(rows []*runtimev1.CanvasRow, keep Keep, prefix string, paths *[]string) []*runtimev1.CanvasRow {
	var out []*runtimev1.CanvasRow
	for i, row := range rows {
		rowPath := fmt.Sprintf("%s.%d", prefix, i)
		if keep.Row != nil && !keep.Row(row) {
			*paths = append(*paths, rowPath)
			continue
		}

		if tg := row.GetTabGroup(); tg != nil {
			var tabs []*runtimev1.CanvasTab
			for t, tab := range tg.Tabs {
				tabPath := fmt.Sprintf("%s.tabs.%d", rowPath, t)
				if keep.Tab != nil && !keep.Tab(tab) {
					*paths = append(*paths, tabPath)
					continue
				}
				tab, ok := pruneNestedRows(tab, keep, tabPath+".rows", paths, func(tab *runtimev1.CanvasTab) *[]*runtimev1.CanvasRow { return &tab.Rows })
				if ok {
					tabs = append(tabs, tab)
				}
			}
			if len(tabs) == 0 {
				continue
			}
			if !slices.Equal(tabs, tg.Tabs) {
				row = proto.Clone(row).(*runtimev1.CanvasRow)
				row.TabGroup.Tabs = tabs
			}
			out = append(out, row)
			continue
		}

		// A container's own `if` is the row's, which keep.Row has already checked.
		if row.GetContainer() != nil {
			row, ok := pruneNestedRows(row, keep, rowPath+".rows", paths, func(row *runtimev1.CanvasRow) *[]*runtimev1.CanvasRow { return &row.Container.Rows })
			if ok {
				out = append(out, row)
			}
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
		if len(items) == 0 {
			continue
		}
		if len(items) != len(row.Items) {
			row = proto.Clone(row).(*runtimev1.CanvasRow)
			row.Items = items
		}
		out = append(out, row)
	}
	if slices.Equal(out, rows) {
		return rows
	}
	return out
}

// pruneNestedRows prunes the rows list that parent holds, such as a tab's rows, at the given YAML path prefix.
// rowsOf returns a pointer to that list for parent or a clone of it.
// It returns false if none of the rows remain, in which case the caller removes parent too.
// Otherwise it returns parent, or a clone of parent with the remaining rows if any were removed.
func pruneNestedRows[T proto.Message](parent T, keep Keep, prefix string, paths *[]string, rowsOf func(T) *[]*runtimev1.CanvasRow) (T, bool) {
	rows := *rowsOf(parent)
	remaining := pruneRows(rows, keep, prefix, paths)
	if len(remaining) == 0 {
		return parent, false
	}
	if !slices.Equal(remaining, rows) {
		parent = proto.Clone(parent).(T)
		*rowsOf(parent) = remaining
	}
	return parent, true
}

// StripConditions clears the `if` conditions of the rows, tabs and items, so they aren't sent to viewers.
// It modifies the rows in place.
func StripConditions(rows []*runtimev1.CanvasRow) {
	WalkRows(rows, Visit{
		Row:  func(row *runtimev1.CanvasRow) { row.ConditionExpression = "" },
		Tab:  func(tab *runtimev1.CanvasTab) { tab.ConditionExpression = "" },
		Item: func(item *runtimev1.CanvasItem) { item.ConditionExpression = "" },
	})
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
	WalkRows(rows, Visit{
		Row:  func(row *runtimev1.CanvasRow) { add(row.ConditionExpression) },
		Tab:  func(tab *runtimev1.CanvasTab) { add(tab.ConditionExpression) },
		Item: func(item *runtimev1.CanvasItem) { add(item.ConditionExpression) },
	})
	return res
}

// Visit holds the functions that WalkRows calls for each kind of canvas element. A nil function isn't called.
type Visit struct {
	// Row is called for plain rows, tab-group rows and container rows.
	Row  func(row *runtimev1.CanvasRow)
	Tab  func(tab *runtimev1.CanvasTab)
	Item func(item *runtimev1.CanvasItem)
}

// WalkRows calls visit for each row, tab and item in rows, including those in tab groups and containers.
// It visits the elements in YAML order, each before the elements it holds.
func WalkRows(rows []*runtimev1.CanvasRow, visit Visit) {
	for _, row := range rows {
		if visit.Row != nil {
			visit.Row(row)
		}
		if visit.Item != nil {
			for _, item := range row.Items {
				visit.Item(item)
			}
		}
		for _, tab := range row.GetTabGroup().GetTabs() {
			if visit.Tab != nil {
				visit.Tab(tab)
			}
			WalkRows(tab.Rows, visit)
		}
		WalkRows(row.GetContainer().GetRows(), visit)
	}
}
