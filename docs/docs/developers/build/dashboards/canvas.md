---
title: Canvas Dashboards
description: Create custom dashboards by assembling visualizations and components
sidebar_label: Canvas Dashboards
sidebar_position: 05
---

While Rill's signature **[Explore dashboards](/developers/build/dashboards/explore)** let you slice, dice, and interact with your data in our suggested layout, **Canvas dashboards** let you define your own visualizations and arrange them into your preferred layout.


Canvas dashboards are built using various components that can display data, create visualizations, and add rich content from independent metrics views. You can create components dynamically through the visual Canvas Dashboard interface or define them in individual YAML files.

## Available Components

Canvas dashboards support three main types of components:

- **[Data components](/developers/build/dashboards/canvas-widgets/data)** - Display structured data and key metrics (KPIs, leaderboards, tables)
- **[Chart components](/developers/build/dashboards/canvas-widgets/chart)** - Create rich visualizations (bar charts, line charts, heat maps, etc.)
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


![Local Filters](/img/build/dashboard/canvas/local-filters.png)

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

## Example Canvas Dashboards

Here are a few deployed examples of Canvas dashboards that you can check out!

- **[E-commerce demo dashboard](https://ui.rilldata.com/demo/ezcommerce-demo/canvas/canvas)**
- **[Programmatic advertising demo dashboard](https://ui.rilldata.com/demo/rill-openrtb-prog-ads/canvas/executive_overview)**
- **[New York City demo dashboard](https://ui.rilldata.com/demo/nyc-canvas-jam/canvas/scorecard%20canvas)**
- **[NYC party demo dashboard 🎉](https://ui.rilldata.com/demo/nyc-canvas-jam/canvas/Leaderboard)**

