package canvas_test

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/canvas"
	"github.com/stretchr/testify/require"
)

func TestPruneRows(t *testing.T) {
	removed := map[string]bool{"cost_table": true, "pipeline_chart": true, "cohorts_chart": true, "leaderboard": true}
	keep := canvas.Keep{Item: func(item *runtimev1.CanvasItem) bool { return !removed[item.Component] }}

	t.Run("nothing removed returns the input", func(t *testing.T) {
		rows := []*runtimev1.CanvasRow{row(item("kpis", 12))}
		res, paths := canvas.PruneRows(rows, keep)
		require.Empty(t, paths)
		require.Same(t, rows[0], res[0])
	})

	t.Run("remaining items keep their widths", func(t *testing.T) {
		rows := []*runtimev1.CanvasRow{
			row(item("margin_chart", 8), item("cost_table", 4)),
			row(item("a", 6), item("b", 3), item("leaderboard", 3)),
		}
		res, paths := canvas.PruneRows(rows, keep)
		require.Equal(t, []string{"rows.0.items.1", "rows.1.items.2"}, paths)
		// The frontend splits these rows equally, since their widths no longer add up to 12
		require.Equal(t, []uint32{8}, widths(res[0]))
		require.Equal(t, []uint32{6, 3}, widths(res[1]))

		// The input is not modified
		require.Len(t, rows[0].Items, 2)
	})

	t.Run("rows without widths are left for the frontend to split", func(t *testing.T) {
		rows := []*runtimev1.CanvasRow{row(&runtimev1.CanvasItem{Component: "a"}, &runtimev1.CanvasItem{Component: "leaderboard"})}
		res, _ := canvas.PruneRows(rows, keep)
		require.Len(t, res[0].Items, 1)
		require.Nil(t, res[0].Items[0].Width)
	})

	t.Run("rows, tabs and groups that lose all their content are removed", func(t *testing.T) {
		rows := []*runtimev1.CanvasRow{
			row(item("kpis", 12)),
			row(item("cost_table", 12)),
			tabGroup("deep_dives",
				tab("pipeline", row(item("pipeline_chart", 12))),
				tab("cohorts", row(item("cohorts_chart", 12))),
			),
			tabGroup("more",
				tab("pipeline", row(item("pipeline_chart", 12))),
				tab("regions", row(item("region_chart", 12))),
				tab("notes"),
			),
		}
		res, paths := canvas.PruneRows(rows, keep)
		require.Equal(t, []string{
			"rows.1.items.0",
			"rows.2.tabs.0.rows.0.items.0",
			"rows.2.tabs.1.rows.0.items.0",
			"rows.3.tabs.0.rows.0.items.0",
		}, paths)
		require.Len(t, res, 2)
		require.Equal(t, "kpis", res[0].Items[0].Component)

		// The group keeps its name and its remaining tab; the tab that was authored empty is removed too
		group := res[1].GetTabGroup()
		require.Equal(t, "more", group.Name)
		require.Len(t, group.Tabs, 1)
		require.Equal(t, "regions", group.Tabs[0].Name)

		// The input is not modified
		require.Len(t, rows[3].GetTabGroup().Tabs, 3)
	})

	t.Run("rows and tabs rejected by their own condition", func(t *testing.T) {
		hidden := func(condition string) bool { return condition == "hide" }
		conditional := canvas.Keep{
			Row:  func(r *runtimev1.CanvasRow) bool { return !hidden(r.ConditionExpression) },
			Tab:  func(t *runtimev1.CanvasTab) bool { return !hidden(t.ConditionExpression) },
			Item: func(i *runtimev1.CanvasItem) bool { return !hidden(i.ConditionExpression) },
		}
		hiddenRow := row(item("margin_chart", 12))
		hiddenRow.ConditionExpression = "hide"
		hiddenTab := tab("pipeline", row(item("pipeline_chart", 12)))
		hiddenTab.ConditionExpression = "hide"
		rows := []*runtimev1.CanvasRow{
			row(item("kpis", 12)),
			hiddenRow,
			tabGroup("deep_dives", hiddenTab, tab("cohorts", row(item("cohorts_chart", 12)))),
		}

		res, paths := canvas.PruneRows(rows, conditional)
		require.Equal(t, []string{"rows.1", "rows.2.tabs.0"}, paths)
		require.Len(t, res, 2)
		require.Equal(t, []string{"cohorts"}, []string{res[1].GetTabGroup().Tabs[0].Name})
		require.Equal(t, []string{"hide"}, canvas.Conditions(rows))

		canvas.StripConditions(res)
		require.Empty(t, canvas.Conditions(res))
	})
}

func item(component string, width uint32) *runtimev1.CanvasItem {
	return &runtimev1.CanvasItem{Component: component, Width: &width}
}

func row(items ...*runtimev1.CanvasItem) *runtimev1.CanvasRow {
	return &runtimev1.CanvasRow{Items: items}
}

func tab(name string, rows ...*runtimev1.CanvasRow) *runtimev1.CanvasTab {
	return &runtimev1.CanvasTab{Name: name, DisplayName: name, Rows: rows}
}

func tabGroup(name string, tabs ...*runtimev1.CanvasTab) *runtimev1.CanvasRow {
	return &runtimev1.CanvasRow{TabGroup: &runtimev1.CanvasTabGroup{Name: name, Tabs: tabs}}
}

func widths(r *runtimev1.CanvasRow) []uint32 {
	var res []uint32
	for _, it := range r.Items {
		res = append(res, *it.Width)
	}
	return res
}
