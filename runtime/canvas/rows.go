package canvas

import (
	"fmt"
	"slices"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/proto"
)

// Keep decides which canvas elements PruneRows keeps. A nil function keeps every element of its kind.
type Keep struct {
	// Row is called for plain rows and for tab-group rows.
	Row  func(row *runtimev1.CanvasRow) bool
	Tab  func(tab *runtimev1.CanvasTab) bool
	Item func(item *runtimev1.CanvasItem) bool
}

// PruneRows returns the rows without the elements that keep rejects,
// and the YAML paths of the rejected elements, such as "rows.1", "rows.0.items.1" or "rows.3.tabs.0.rows.0.items.0".
// Rows, tabs and tab groups left without content are removed too.
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
				tabRows := pruneRows(tab.Rows, keep, tabPath+".rows", paths)
				if len(tabRows) == 0 {
					continue
				}
				if !slices.Equal(tabRows, tab.Rows) {
					tab = proto.Clone(tab).(*runtimev1.CanvasTab)
					tab.Rows = tabRows
				}
				tabs = append(tabs, tab)
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
