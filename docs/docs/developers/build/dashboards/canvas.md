---
title: Build Canvas Dashboards
description: Create custom dashboards by assembling visualizations and components
sidebar_label: Canvas Dashboards
sidebar_position: 05
---

While Rill's signature **[Explore dashboards](/developers/build/dashboards/explore)** let you slice, dice, and interact with your data in our suggested layout, **Canvas dashboards** let you define your own visualizations and arrange them into your preferred layout.


Canvas dashboards are built using various components that can display data, create visualizations, and add rich content from independent metrics views. You can create components dynamically through the visual Canvas Dashboard interface or define them in individual YAML files.

## Available Components

Canvas dashboards support five main types of components:

- **[Data components](/developers/build/dashboards/canvas-widgets/data)** - Display structured data and key metrics (KPIs, leaderboards, tables)
- **[Chart components](/developers/build/dashboards/canvas-widgets/chart)** - Create rich visualizations (bar charts, line charts, heat maps, etc.)
- **[Map components](/developers/build/dashboards/canvas-widgets/map)** - Plot measures by location as points or shaded regions
- **[Custom viz components](/developers/build/dashboards/canvas-widgets/custom-viz)** - Reusable charts with params, built from 32 chart types or a Vega-Lite spec
- **[Miscellaneous components](/developers/build/dashboards/canvas-widgets/misc)** - Add text, images, and other content

Each component page shows both the visual result and the corresponding YAML code, making it easy to understand how to implement them in your own dashboards.

For a complete overview of all available components, see our [**Canvas components**](/developers/build/dashboards/canvas-widgets) reference.

## Creating a Canvas Dashboard
### A Visual Editing Experience 

To modify any single widget, click to select it and use the right-hand panel to change its associated properties. Click outside the widget to view global properties associated with the overall Canvas.

![Selected Widget](/img/build/dashboard/canvas/selected-widget.png)


### Filters
Optionally toggle on the **global filter bar** under Canvas properties to give dashboard viewers access to the same time and dimension filters available on Explore dashboards.

![Global Filter Bar](/img/build/dashboard/canvas/global-filter-bar.png)

**Local filters** for a single KPI, Chart, or Table can be separated from the global filters via the "Filters" tab in the properties panel, where you can set filters that are local to just that widget.

Local filters are shown as chips in the widget's header. To keep the filter but hide the chips from viewers, turn off **Show filters on component** in the same "Filters" tab (or set `hide_local_filters: true` in YAML).

To keep a filter in the global filter bar at all times, or to block the canvas until viewers select a value for it, use `filters.pinned` and `filters.required`. See [Pin and require filters](/guide/dashboards/canvas#pin-and-require-filters).


![Local Filters](/img/build/dashboard/canvas/local-filters.png)

### Tab Groups

A tab group shows several sets of rows in one place, one tab at a time. To add one, choose **Tab group** in the add menu between rows, or select **Convert row to tab group** in a component's menu. In YAML, a tab group is an entry in `rows` with `tabs` instead of `items`. See the [`rows`](/reference/project-files/canvas-dashboards#rows) reference.

You can move or copy a whole tab group in the visual editor:

- **Move:** Drag the grip handle at the left of the tab strip, or an empty part of the strip, and drop the group between two rows. A tab group can't be dropped inside another tab group or a container.
- **Duplicate:** Open any tab's three-dot menu and select **Duplicate tab group**, or select the tab group and click **Duplicate tab group** in the right-hand panel. The copy, with all its tabs and their rows, is added right below the original and selected.

![A tab's menu with the Duplicate tab group option](/img/build/dashboard/canvas/tab-menu-duplicate-tab-group.png)

![The tab group panel with the Duplicate tab group and Delete tab group buttons](/img/build/dashboard/canvas/tab-group-inspector.png)

The copy gets a name of its own, such as `group-2`, which is its deep-link key in the dashboard URL. You can change it in the **Tab group name** field.

### Containers

A container keeps the rows of one story together, so you can move, copy, delete, and show or hide them as one unit. Viewers don't see the container itself: its rows look like any other rows. In the visual editor, a dashed frame with a **Container** label marks it.

In YAML, a container is an entry in `rows` with its own `rows` instead of `items`:

```yaml
rows:
  - name: revenue_story # optional; defaults to container-<index>
    rows:
      - items:
          - kpi_grid:
              metrics_view: orders
              measures: [revenue, margin]
      - items:
          - width: 8
            line_chart: # ...
          - width: 4
            table: # ...
```

Containers are top-level entries that hold plain rows: they can't be placed inside a tab group or another container, and they can't hold tab groups. See the [`rows`](/reference/project-files/canvas-dashboards#rows) reference.

You can also manage containers in the visual editor:

- **Add:** Choose **Container** in the add menu between rows, or select **Wrap row in container** in a component's menu. These options appear when the `canvas_containers` [feature flag](/developers/build/project-configuration#feature-flags) is on.
- **Fill:** Drag widgets into the container, or click **Add widget to container** in an empty one. Drag them out the same way.
- **Move:** Drag the grip handle on the container's label and drop the container between two rows.
- **Rename:** Click the container's label and change **Container name** in the right-hand panel.
- **Duplicate, ungroup or delete:** Open the three-dot menu on the label, or select the container and use the right-hand panel. **Ungroup** removes the container and keeps its rows in place.

The visual editor writes a `name` for every tab group and container. A new one gets a name such as `container-2`. When you edit the layout, each unnamed one gets the name it had by default, such as `group-3`, which comes from its position. A block then keeps its name, and a tab group its deep-link key, wherever you move it.

### Making changes to the YAML 
While we encourage creating Canvas dashboards via the visual editing experience described above, you can always edit the YAML file directly using the code view by toggling the switch next to the filename at the top of the page. Please see our [customization page](/developers/build/dashboards/customization) and [reference documentation](/reference/project-files/canvas-dashboards) for more information.


![Code Toggle](/img/build/dashboard/canvas/code-toggle.png)

:::tip Customize default time ranges
Set project-wide default time ranges and available options for all canvas dashboards.
[Learn more about canvas defaults →](/developers/build/project-configuration#canvas-defaults)
::: 

## Default Filters

Dashboard creators can configure default filters to establish a consistent starting point for viewers. Filters are defined as Metrics SQL WHERE expressions, keyed by the metrics view name they apply to.

```yaml
defaults:
  filters:
    # Key is the metrics view name; value is a Metrics SQL WHERE expression
    my_metrics_view: "country IN ('US', 'CA') AND revenue > 1000"
    another_metrics_view: "status = 'active'"
```

This lets you pre-filter data across one or more metrics views used in the canvas, ensuring users begin their analysis with the most relevant context.

For detailed YAML configurations, see the [`defaults`](/reference/project-files/canvas-dashboards#defaults) section in our reference documentation.

## Show Content to Some Viewers Only

Add an `if` condition to a row, tab group, tab, container or component to show it only to some viewers. One canvas can then serve several audiences, such as finance and sales teams, or the basic and premium tiers of an embedded dashboard.

```yaml
rows:
  # Everyone sees the KPIs
  - items:
      - kpi_grid:
          metrics_view: orders
          measures: [total_revenue, order_count]

  # Only members of the finance group see this row
  - if: '{{ has "finance" .user.groups }}'
    items:
      - width: 8
        line_chart:
          metrics_view: orders
          x: { field: order_date, type: temporal }
          y: { field: gross_margin, type: quantitative }
      # Only admins in the finance group see this table
      - width: 4
        if: '{{ .user.admin }}'
        table:
          metrics_view: costs
          columns: [cost_center, total_cost]

  - name: deep_dives
    tabs:
      - label: Pipeline
        if: '{{ has "sales" .user.groups }}'
        rows: # ...
      - label: Cohorts
        # `plan` is a custom attribute, e.g. passed by an embedding app
        if: '{{ eq .user.plan "premium" }}'
        rows: # ...
```

Conditions use the same templating and [user attributes](/developers/build/metrics-view/security#user-attributes) as security policies. A few things to know:

- An element shows only when its own condition and the conditions of the rows, tabs, tab groups and containers holding it are all true. A container's condition is the simplest way to show or hide a whole story.
- Conditions can't reference project variables (`.env`), or the full template data (`.` or `$`) that includes them, since they are evaluated for viewers and the variables commonly hold secrets. The same applies to all canvas component properties, such as Markdown content and chart SQL.
- Built-in attributes (`name`, `email`, `domain`, `groups`, `admin` and `embed`) always resolve, even for an [embedded dashboard](/developers/embed/iframe) that only passes custom attributes. A condition that references a custom attribute the viewer doesn't have is false. To treat a missing attribute as a value instead, use `get`, as in `'{{ ne (get .user "plan") "premium" }}'`.
- When components are hidden, the rest of the row widens to fill it, and rows, tabs and containers left empty disappear.
- Hidden content is removed before the dashboard reaches the viewer, including its components and the conditions themselves. In an embedded dashboard or a public URL that's restricted to the canvas, the viewer also can't query the metrics views that only hidden content uses.
- Hiding content doesn't restrict data elsewhere: a viewer with access to a metrics view can still query it, for example in an Explore dashboard. To restrict data, add a [security policy](/developers/build/metrics-view/security) to the metrics view.
- Public URLs and scheduled reports that run as their creator show what the creator can see.

### Set Conditions in the Visual Editor

To set a component's condition without editing YAML, select the component and fill in **Visible when** at the bottom of the right-hand panel. To set the condition of a tab group or a container, select it and fill in **Visible when** in its panel. Leave a field empty to show the element to everyone. Conditions on rows and tabs are set in YAML.

![The Visible when field in the component panel, set to an admin-only condition](/img/build/dashboard/canvas/visible-when-field.png)

In the editor, you always see the whole canvas. Components that are only shown to some viewers have a crossed-out eye badge in their bottom-right corner. Hover over it to see the conditions that apply, including those inherited from the component's row, tab, tab group or container. A container with a condition shows the same badge on its label.

![A crossed-out eye badge on a canvas component, with a tooltip showing its partners-only condition](/img/build/dashboard/canvas/visibility-badge.png)

### Check What Each Audience Sees

Add [mock users](/developers/build/metrics-view/view-as-user) with the attributes you want to test to `rill.yaml`:

```yaml
mock_users:
  - email: someone@rilldata.com
    name: John Doe
    admin: true
  - email: jane@partnercompany.com
    name: Jane Partner
    groups:
      - partners
```

Then open the dashboard preview, click **View as**, and pick a mock user. **View as** appears when the canvas has `if` conditions, or when a security policy applies to the canvas or its metrics views.

![The View as menu listing two mock users](/img/build/dashboard/canvas/view-as-mock-users.png)

For example, this canvas has a welcome row with `if: '{{ has "partners" .user.groups }}'` and a heatmap with `if: '{{ .user.admin }}'`. Viewing as Jane, who is in the `partners` group, shows the welcome row but not the heatmap. Her KPI grid also has no spend tile, because the metrics view's [security policy](/developers/build/metrics-view/security) hides that measure from partners.

![The canvas viewed as Jane: the partner welcome row and three KPIs, without the heatmap](/img/build/dashboard/canvas/view-as-partner.png)

Viewing as John, an admin outside the `partners` group, shows the heatmap but not the welcome row.

![The canvas viewed as John: four KPIs and the admin-only heatmap, without the partner welcome row](/img/build/dashboard/canvas/view-as-admin.png)

## Example Canvas Dashboards

Here are a few deployed examples of Canvas dashboards that you can check out!

- **[E-commerce demo dashboard](https://ui.rilldata.com/demo/ezcommerce-demo/canvas/canvas)**
- **[Programmatic advertising demo dashboard](https://ui.rilldata.com/demo/rill-openrtb-prog-ads/canvas/executive_overview)**
- **[New York City demo dashboard](https://ui.rilldata.com/demo/nyc-canvas-jam/canvas/scorecard%20canvas)**
- **[NYC party demo dashboard 🎉](https://ui.rilldata.com/demo/nyc-canvas-jam/canvas/Leaderboard)**

