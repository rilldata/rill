---
title: Dashboard URL Parameters
description: Reference guide for all URL parameters used in Rill dashboard stateful URLs
sidebar_label: URL Parameters
sidebar_position: 12
---

# Dashboard URL Parameters

Rill dashboards use stateful URLs that encode the current dashboard state in the URL query parameters. This allows you to bookmark specific views, share links with others, and programmatically construct URLs to navigate to specific dashboard configurations.

## Overview

All dashboard state is encoded in the URL query string. As you interact with a dashboard (change filters, time ranges, views, etc.), the URL automatically updates to reflect the current state. This makes it easy to:

- **Share specific views**: Copy the URL to share an exact dashboard configuration
- **Bookmark views**: Save URLs to return to specific dashboard states
- **Programmatic control**: Construct URLs programmatically to navigate to specific configurations
- **Deep linking**: Link directly to specific dashboard states from other applications

## URL Parameter Reference

### View Parameters

#### `view`

Specifies which dashboard view is currently active.

**Values:**
- `explore` - The default explore view (leaderboard/table view)
- `pivot` - Pivot table view
- `tdd` - Time dimension detail view (time series chart view)

**Example:**
```
?view=pivot
```

---

### Time Parameters

#### `tr` (Time Range)

Specifies the primary time range for the dashboard.

**Format:**
- **Preset ranges**: Use preset names like `P7D` (last 7 days), `P30D` (last 30 days), `P1M` (last month), etc.
- **Custom ranges**: Use ISO 8601 format: `2024-01-01T00:00:00.000Z,2024-01-31T23:59:59.999Z`
- **Duration-based**: Use Rill time syntax like `7d as of now/d` (7 days ending at start of current day)
- **Fixed-point**: Use syntax like `-7d/d to now/d` (from start of day 7 days ago to start of current day)

**Example:**
```
?tr=P7D
?tr=2024-01-01T00:00:00.000Z,2024-01-31T23:59:59.999Z
```

#### `tz` (Time Zone)

Specifies the timezone for time-based operations.

**Format:** IANA timezone identifier (e.g., `UTC`, `America/New_York`, `Europe/London`)

**Example:**
```
?tz=America/New_York
```

#### `grain` (Time Grain)

Specifies the time granularity for aggregating time series data.

**Values:** `second`, `minute`, `hour`, `day`, `week`, `month`, `quarter`, `year`

**Example:**
```
?grain=day
```

#### `compare_tr` (Comparison Time Range)

Specifies a comparison time range for time-based comparisons.

**Format:** Same as `tr` parameter. Can be preset names (e.g., `rill-PP` for previous period) or custom ranges.

**Example:**
```
?compare_tr=rill-PP
```

#### `compare_dim` (Comparison Dimension)

Specifies which dimension to use for comparison when comparing across dimensions.

**Format:** Dimension name from your metrics view

**Example:**
```
?compare_dim=region
```

#### `highlighted_tr` (Highlighted Time Range)

Specifies a highlighted/scrubbed time range, typically set by interacting with the time series chart.

**Format:** Same as `tr` parameter (ISO 8601 format for custom ranges)

**Example:**
```
?highlighted_tr=2024-01-15T00:00:00.000Z,2024-01-20T23:59:59.999Z
```

---

### Filter Parameters

#### `f` (Filters)

Specifies filter expressions to apply to the dashboard data.

**Format:** URL-encoded filter expression using Rill's filter syntax. Supports:
- Comparison operators: `IN`, `NOT IN`, `LIKE`, `NOT LIKE`
- Logical operators: `AND`
- Column names and values

**Example:**
```
?f=region+IN+%28%27North%27%2C%27South%27%29
```

**Note:** Filter expressions are URL-encoded. The decoded version of the first example would be:
```
region IN ('North','South')
```

---

### Explore View Parameters

These parameters apply when `view=explore` (or when no view is specified, as explore is the default).

#### `measures`

Specifies which measures are visible in the explore view.

**Format:** Comma-separated list of measure names

**Example:**
```
?measures=revenue,orders,users
```

#### `dims` (Dimensions)

Specifies which dimensions are visible in the explore view.

**Format:** Comma-separated list of dimension names

**Example:**
```
?dims=region,category,status
```

#### `expand_dim` (Expanded Dimension)

Specifies which dimension is currently expanded in the leaderboard view.

**Format:** Dimension name

**Example:**
```
?expand_dim=region
```

#### `sort_by`

Specifies which field to sort by in the leaderboard view.

**Format:** Measure name or dimension name

**Example:**
```
?sort_by=revenue
```

#### `sort_type`

Specifies the type of sorting to apply.

**Values:**
- `value` - Sort by the actual value
- `percent` - Sort by percentage
- `delta_abs` - Sort by absolute delta (change)
- `delta_percent` - Sort by percentage delta
- `dim` - Sort by dimension value

**Example:**
```
?sort_type=delta_percent
```

#### `sort_dir` (Sort Direction)

Specifies the sort direction.

**Values:** `ASC` (ascending) or `DESC` (descending)

**Example:**
```
?sort_dir=DESC
```

#### `leaderboard_measures`

Specifies which measures are shown in the leaderboard context.

**Format:** Comma-separated list of measure names

**Example:**
```
?leaderboard_measures=revenue,orders
```

#### `lb_ctx` (Leaderboard Show Context For All Measures)

Boolean flag indicating whether to show context for all measures in the leaderboard.

**Values:** `true` or `false`

**Example:**
```
?lb_ctx=true
```

---

### Time Dimension Detail View Parameters

These parameters apply when `view=tdd`.

#### `measure` (Expanded Measure)

Specifies which measure is currently expanded in the time dimension detail view.

**Format:** Measure name

**Example:**
```
?measure=revenue
```

#### `chart_type`

Specifies the chart type for the time series visualization.

**Values:**
- `timeseries` or `line` - Line chart (default)
- `bar` - Grouped bar chart
- `stacked_bar` - Stacked bar chart
- `stacked_area` - Stacked area chart

**Example:**
```
?chart_type=stacked_bar
```

#### `pin`

Boolean flag indicating whether the time dimension detail view is pinned.

**Values:** `true` or `false` (or just present/absent)

**Example:**
```
?pin=true
```

---

### Pivot View Parameters

These parameters apply when `view=pivot`.

#### `rows`

Specifies which dimensions/fields are used as rows in the pivot table.

**Format:** Comma-separated list of dimension names or time dimensions (e.g., `time.day`, `time.month`)

**Example:**
```
?rows=region,category
?rows=time.month,region
```

#### `cols` (Columns)

Specifies which dimensions/measures are used as columns in the pivot table.

**Format:** Comma-separated list of dimension names, measure names, or time dimensions

**Example:**
```
?cols=revenue,orders
?cols=time.year,status
```

#### `table_mode`

Specifies the pivot table display mode.

**Values:**
- `nest` - Nested table mode (default)
- `flat` - Flat table mode

**Example:**
```
?table_mode=flat
```

#### `row_limit`

Specifies the maximum number of rows to display in the pivot table.

**Format:** Integer number

**Example:**
```
?row_limit=100
```

**Note:** When in pivot view, `sort_by` and `sort_dir` can also be used to control pivot table sorting.

---

### Advanced Parameters

#### `gzipped_state`

Contains a gzipped and base64-encoded version of the full dashboard state. This is used when the URL would otherwise become too long.

**Format:** Base64-encoded gzipped state string

**Note:** This parameter is automatically used by Rill when the URL would exceed reasonable length limits. You typically don't need to construct this manually.

#### `state` (Legacy Proto State)

Legacy parameter for backward compatibility. Contains a protobuf-encoded dashboard state.

**Note:** This parameter is maintained for backward compatibility with older Rill URLs. New URLs should use the individual parameters instead.

---

## Complete URL Examples

### Explore View with Filters and Time Range

```
https://ui.rilldata.com/org/project/explore/dashboard-name?view=explore&tr=P7D&grain=day&f=region+IN+%28%27North%27%29&measures=revenue,orders&expand_dim=region&sort_by=revenue&sort_dir=DESC
```

### Pivot Table View

```
https://ui.rilldata.com/org/project/explore/dashboard-name?view=pivot&tr=P30D&rows=region,category&cols=revenue,orders&table_mode=nest&sort_by=revenue&sort_dir=DESC
```

### Time Dimension Detail View

```
https://ui.rilldata.com/org/project/explore/dashboard-name?view=tdd&tr=P7D&grain=hour&measure=revenue&chart_type=stacked_area&compare_tr=rill-PP
```

### Complex Example with Multiple Parameters

```
https://ui.rilldata.com/org/project/explore/dashboard-name?view=explore&tr=P30D&tz=America/New_York&grain=day&compare_tr=rill-PP&compare_dim=region&f=status+%3D+%27active%27&measures=revenue,orders,users&dims=region,category&expand_dim=region&sort_by=revenue&sort_type=delta_percent&sort_dir=DESC&leaderboard_measures=revenue,orders&lb_ctx=true
```

---

## URL Encoding

When constructing URLs programmatically, ensure that:

1. **Special characters are URL-encoded**: Spaces become `+` or `%20`, parentheses become `%28` and `%29`, etc.
2. **Filter expressions are properly encoded**: Use `encodeURIComponent()` or similar functions
3. **Comma-separated lists are not encoded**: Lists like `measures=revenue,orders` should not have commas encoded

### JavaScript Example

```javascript
const params = new URLSearchParams();
params.set('view', 'explore');
params.set('tr', 'P7D');
params.set('grain', 'day');
params.set('f', "region IN ('North','South')"); // Will be auto-encoded
params.set('measures', 'revenue,orders'); // Commas don't need encoding

const url = `https://ui.rilldata.com/org/project/explore/dashboard-name?${params.toString()}`;
```

---


## Integration with Embed API

When using the [Embed API](/developers/embed/embed-iframe-api), you can use these URL parameters in the `setState` method:

```javascript
iframe.contentWindow.postMessage({
  id: 1,
  method: "setState",
  params: "view=explore&tr=P7D&grain=day&measures=revenue,orders"
}, "*");
```

The state string passed to `setState` should use the same format as the URL query string (without the leading `?`).

---

## See Also

- [Embed Dashboards](/developers/embed/embedding) - Learn how to embed Rill dashboards
- [Embed Iframe API](/developers/embed/embed-iframe-api) - Programmatically control embedded dashboards
- [Time Series Filter]/guide/dashboards/time-series/time-series) - Detailed guide on time range syntax and usage

