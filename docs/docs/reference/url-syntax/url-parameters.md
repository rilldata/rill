---
title: Dashboard URL Parameters
description: Reference for all URL query parameters that control Rill dashboard state.
sidebar_label: URL Parameters
sidebar_position: 10
---

# Dashboard URL Parameters

Rill dashboards encode their full state in the URL query string. These parameters can be used in browser URLs, shared links, bookmarks, and the [Embed Iframe API](/developers/embed/iframe-api) `setState` method.

## Global Parameters

| Parameter | Values / Format | Description |
|---|---|---|
| `view` | `explore` (default), `pivot`, `tdd` | Active dashboard view |
| `tr` | ISO 8601 duration (`P7D`, `P30D`) or range (`2024-01-01T00:00:00.000Z,2024-01-31T23:59:59.999Z`) | Time range. Also supports Rill syntax (`-7d/d to now/d`) |
| `tz` | IANA identifier (`UTC`, `America/New_York`) | Timezone |
| `grain` | `second`, `minute`, `hour`, `day`, `week`, `month`, `quarter`, `year` | Time aggregation granularity |
| `compare_tr` | Same format as `tr`. Use `rill-PP` for previous period | Comparison time range |
| `compare_dim` | Dimension name | Dimension for comparison |
| `highlighted_tr` | ISO 8601 range | Highlighted/scrubbed time range on the time series chart |
| `f` | URL-encoded filter expression | Filters. Decoded example: `region IN ('North','South')` |

## Explore View Parameters

Apply when `view=explore` or no view is set.

| Parameter | Values / Format | Description |
|---|---|---|
| `measures` | Comma-separated measure names | Visible measures |
| `dims` | Comma-separated dimension names | Visible dimensions |
| `expand_dim` | Dimension name | Expanded dimension in leaderboard |
| `sort_by` | Measure or dimension name | Sort field |
| `sort_type` | `value`, `percent`, `delta_abs`, `delta_percent`, `dim` | Sort metric type |
| `sort_dir` | `ASC`, `DESC` | Sort direction |
| `leaderboard_measures` | Comma-separated measure names | Measures shown in leaderboard context |
| `lb_ctx` | `true`, `false` | Show context for all leaderboard measures |

## Time Dimension Detail Parameters

Apply when `view=tdd`.

| Parameter | Values / Format | Description |
|---|---|---|
| `measure` | Measure name | Expanded measure |
| `chart_type` | `timeseries` / `line`, `bar`, `stacked_bar`, `stacked_area` | Chart visualization type |
| `pin` | `true`, `false` | Pin the detail view |

## Pivot View Parameters

Apply when `view=pivot`. The `sort_by` and `sort_dir` global parameters also apply here.

| Parameter | Values / Format | Description |
|---|---|---|
| `rows` | Comma-separated dimension names or time dims (`time.day`, `time.month`) | Row fields |
| `cols` | Comma-separated dimension, measure, or time dim names | Column fields |
| `table_mode` | `nest` (default), `flat` | Table display mode |
| `row_limit` | Integer | Maximum rows displayed |

## Advanced Parameters

| Parameter | Description |
|---|---|
| `gzipped_state` | Base64-encoded gzipped state. Auto-generated when URL length exceeds limits. |
| `state` | Legacy protobuf-encoded state. Maintained for backward compatibility; prefer individual parameters. |

## URL Encoding

When constructing URLs programmatically:
- URL-encode special characters in filter expressions (`%28`, `%29`, `+` for spaces)
- Do **not** encode commas in list parameters (`measures=revenue,orders`)

```javascript
const params = new URLSearchParams();
params.set("view", "explore");
params.set("tr", "P7D");
params.set("grain", "day");
params.set("f", "region IN ('North','South')");
params.set("measures", "revenue,orders");

const url = `https://ui.rilldata.com/org/project/explore/dashboard-name?${params}`;
```

