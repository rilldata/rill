---
title: "Exact Values"
description: Switch an approximate measure to its exact value with one click
sidebar_label: "Exact Values"
sidebar_position: 36
---

Some measures are expensive to compute exactly. Distinct counts such as unique users, and quantiles such as P95 latency, are therefore often defined with a fast approximate aggregate that is within a percent or two of the truth. When the project developer also defines an [exact expression](/developers/build/metrics-view/measures/exact-measures) for such a measure, the dashboard shows the approximate value by default and lets you switch to the exact value when you need to reconcile a number.

## Read the chip

A small chip after the measure name tells you how the value was computed:

| Chip | Meaning |
|------|---------|
| `≈` outlined | Approximate. The default; fast. |
| `=` filled | Exact. You switched this measure to its exact expression. |
| spinner | The exact value is being computed. The previous approximate value stays visible, dimmed, until it arrives. |
| `!` | The exact query failed or timed out. Hover for the error, click to return to the approximate value. |

Measures without an exact expression show no chip. That includes measures calculated from other measures, such as a ratio of two unique counts: they are computed from the values shown next to them, so a ratio follows the precision of the unique counts in the same table.

Each chip controls the part of the dashboard it sits in: the big-number row (the big numbers and their charts), one leaderboard, the dimension table, the Time Dimension Detail page (its chart and table) or the pivot. Within that part, precision is resolved once, so a measure never has two values in the same table.

## Switch a measure to exact

Click the `≈` chip next to the measure name, in a big number, a leaderboard or dimension table header, the Time Dimension Detail header, or a pivot column header. Only that part of the dashboard is recomputed exactly: the big-number row with its charts, that one leaderboard, the dimension table, the Time Dimension Detail page or the pivot. Other leaderboards and tables keep their approximate value until you click their own chip, so each click runs one expensive query rather than one for every table on the page. Click the chip again to return to the approximate value.

Exact queries can take noticeably longer than approximate ones, so the switch applies only to the view you are looking at:

- Changing the time range, comparison, grain, timezone, filters, the selected dimension or the view (explore, dimension table, Time Dimension Detail, pivot) returns every measure to approximate. New queries are always approximate until you click again.
- Showing, hiding or reordering measures keeps your selection.
- The selection is not part of the URL, so bookmarks, shared links and public URLs always open approximate. Scheduled reports and alerts are approximate as well.
- A CSV or Parquet export downloads what you see, so exporting while a measure is exact gives you the exact values.
- A measure filter (for example "publishers where unique users > 1,000") is evaluated with the exact value only in queries that also show that measure. Switch the measure to exact and keep it visible when the filter has to be exact.

In a canvas dashboard the chip works the same way for KPIs, tables, pivots, leaderboards and charts. A click applies to that widget only, and to one leaderboard at a time inside a leaderboard widget. KPIs and table, pivot and leaderboard columns each have their own chip. A chart has one chip, next to its title, that switches every measure of the chart at once.
