---
description: Detailed instructions and examples for developing explore dashboard resources in Rill
---

# Instructions for developing an explore dashboard in Rill

## Introduction

Explore dashboards are resources that configure an interactive, drill-down dashboard for a metrics view. They are Rill's default dashboard type, designed for explorative slice-and-dice analysis of a single metrics view.

Explore dashboards are lightweight resources that sit downstream of a metrics view in the project DAG. Their reconcile logic is fast (validation only), so they can be created and modified freely without performance concerns.

### When to use explores vs canvases

- **Explore dashboards:** Best for explorative analysis, drill-down investigations, and letting users freely slice data by any dimension.
- **Canvas dashboards:** Best for fixed reports, executive summaries, or combining multiple metrics views into a single view.

## Development approach

**Check for an existing inline explore first.** Most metrics views already emit an explore dashboard: any metrics view with `version: 1` and an `explore:` block, and any legacy metrics view without a `version:` property. If the target metrics view already emits an explore, do NOT create a stand-alone `type: explore` file for it; that produces a second, duplicate dashboard. Instead, edit the `explore:` block in the metrics view file. If you are a sub-agent restricted to a different path, report this back to the parent agent instead of writing a duplicate file.

Only create a stand-alone explore file when the metrics view needs more than one explore dashboard, or when the user explicitly asks for a separate file.

Explore dashboards require minimal configuration. In most cases, you only need to:

1. Reference the metrics view
2. Select which dimensions and measures to expose (usually all, indicated by `'*'`)
3. Optionally configure defaults and time ranges

**Best practice:** Keep explore configurations simple. Only add advanced features (security policies, custom themes, restricted dimensions) when there is a clear requirement. The metrics view already defines the business logic; the explore just controls presentation and access.

## Inline explores in metrics views

The default way to create an explore is inline in the metrics view file: set `version: 1` and add an `explore:` block, which emits an explore resource with the same name as the metrics view (or `name:` if set):

```yaml
# metrics/sales.yaml
version: 1
type: metrics_view
display_name: Sales Analytics

model: sales_model
timeseries: order_date

dimensions:
  - column: region
  - column: product_category

measures:
  - name: total_revenue
    expression: SUM(revenue)

# Inline explore configuration
explore:
  display_name: Sales Dashboard
  dimensions: '*'  # Optional: dimensions to expose ('*', a list, or {exclude: [...]}); defaults to all
  measures: '*'    # Optional: measures to expose ('*', a list, or {exclude: [...]}); defaults to all
  time_ranges:
    - P7D
    - P30D
    - P90D
  defaults:
    time_range: P30D
```

For legacy reasons, metrics views without `version:` auto-emit an explore even without an `explore:` block; metrics views with `version: 1` only emit one when the block is present.

Use inline explores to keep the metrics view and its dashboard configuration together. Use separate explore files only when you need multiple explores for the same metrics view (for example, a second dashboard that exposes a restricted subset of dimensions), and give each one a distinct name so it does not collide with the inline explore.

## Stand-alone explore example with annotations

The examples below are for stand-alone explore files. Only use them in the cases described above; otherwise put the same properties in the metrics view's `explore:` block.

Note that most explore dashboards work great without any of the optional properties shown here.

```yaml
# dashboards/sales_explore.yaml

# Required: resource type
type: explore

# Required: the metrics view this dashboard renders
metrics_view: sales_metrics

# Optional: display name shown in the navigation and header
display_name: "Sales Performance"

# Optional: informational banner at the top of the dashboard
banner: "Data refreshes daily at 6 AM UTC"

# Optional: which dimensions to expose (use '*' for all)
dimensions: '*'

# Optional: which measures to expose (use '*' for all)
measures: '*'

# Optional: customize the time range dropdown
time_ranges:
  - P7D
  - P30D
  - P90D
  - P12M
  - rill-MTD
  - rill-YTD

# Optional: default dashboard state on first load
defaults:
  time_range: P30D
  comparison_mode: time

# Optional: pin specific time zones to the top of the selector
time_zones:
  - America/Los_Angeles
  - America/New_York

# Optional: custom theme
theme: my_theme

# Optional: restrict access to specific users.
# Note: usually you should do this in the metrics view, not the explore resource.
security:
  access: "{{ .user.admin }} OR '{{ .user.email }}' LIKE '%@example.com'"
```

## Minimal stand-alone example

When a stand-alone file is warranted, a minimal explore is sufficient:

```yaml
type: explore
metrics_view: sales_metrics
display_name: "Sales Dashboard"
dimensions: '*'
measures: '*'
```

## Reference documentation

Here is a full JSON schema for the explore syntax:

```
{% json_schema_for_resource "explore" %}
```
