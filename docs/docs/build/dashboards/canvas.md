---
title: Canvas Dashboards
description: Create custom dashboards by assembling visualizations and components
sidebar_label: Canvas Dashboards
sidebar_position: 05
---

While Rill's signature **[Explore dashboards](/build/dashboards/explore)** let you slice, dice, and interact with your data in our suggested layout, **Canvas dashboards** let you define your own visualizations and arrange them into your preferred layout.


Canvas dashboards are built using various components that can display data, create visualizations, and add rich content from independent metrics views. You can create components dynamically through the visual Canvas Dashboard interface or define them in individual YAML files.

## Available Components

Canvas dashboards support three main types of components:

- **[Data components](/build/dashboards/canvas-widgets/data)** - Display structured data and key metrics (KPIs, leaderboards, tables)
- **[Chart components](/build/dashboards/canvas-widgets/chart)** - Create rich visualizations (bar charts, line charts, heat maps, etc.)
- **[Miscellaneous components](/build/dashboards/canvas-widgets/misc)** - Add text, images, and other content

Each component page shows both the visual result and the corresponding YAML code, making it easy to understand how to implement them in your own dashboards.

For a complete overview of all available components, see our [**Canvas components**](/build/dashboards/canvas-widgets) reference.

## Creating a Canvas Dashboard
### A Visual Editing Experience 

To modify any single widget, click to select it and use the right-hand panel to change its associated properties. Click outside the widget to view global properties associated with the overall Canvas.

<img src = '/img/build/dashboard/canvas/selected-widget.png' class='rounded-gif' />
<br/>


### Filters
Optionally toggle on the **global filter bar** under Canvas properties to give dashboard viewers access to the same time and dimension filters available on Explore dashboards.

<img src = '/img/build/dashboard/canvas/global-filter-bar.png' class='rounded-gif' />
<br/>

**Local filters** for a single KPI, Chart, or Table can be separated from the global filters via the "Filters" tab in the properties panel, where you can set filters that are local to just that widget.


<img src = '/img/build/dashboard/canvas/local-filters.png' class='rounded-gif' />
<br/>

### Making changes to the YAML 
While we encourage creating Canvas dashboards via the visual editing experience described above, you can always edit the YAML file directly using the code view by toggling the switch next to the filename at the top of the page. Please see our [customization page](/build/dashboards/customization) and [reference documentation](/reference/project-files/canvas-dashboards) for more information.


<img src = '/img/build/dashboard/canvas/code-toggle.png' class='rounded-gif' />
<br/>

:::tip Customize default time ranges
Set project-wide default time ranges and available options for all canvas dashboards.
[Learn more about canvas defaults →](/build/project-configuration#canvas-defaults)
::: 

<!-- 
## Default Filters

Dashboard creators can configure default dimension and measure filters to establish a consistent starting point for viewers. This feature enables you to pre-configure the most relevant data views, ensuring users begin their analysis with the most appropriate context and reducing the need for manual filter configuration.

### Dimension Filters

<img src = '/img/build/dashboard/canvas/dimension-default-filters.png' class='rounded-gif' /> <br/>

Additional parameters can be configured to control filter behavior, including locking filters, hiding values, and setting default selections.

```yaml
defaults:
  filters:
    dimensions:
      # Exclude "Not Available" values and lock the filter
      - dimension: app_site_domain
        exclude: true
        locked: true
        values:
          - "Not Available"
      
      # Make filter non-removable but allow value changes
      - dimension: app_site_name
        removable: false
      
      # Standard dimension filter
      - dimension: device_state
```

### Measure Filters

Measure filters can be configured in the same way:

```yaml
defaults:
  filters:
    measures:
      # Lock impressions filter with greater than 10 threshold
      - measure: impressions
        locked: true
        # hidden: true  # Uncomment to hide from UI
        by_dimension: app_site_domain
        operator: gt  # Available: gt, lt, gte, lte, bt, nbt, eq, neq
        values:
          - "10"
      
      # Configure total_bids filter with less than or equal to 10
      - measure: total_bids
        by_dimension: app_site_name
        operator: lte  # Available: gt, lt, gte, lte, bt, nbt, eq, neq
        values:
          - "10"
```

<img src = '/img/build/dashboard/canvas/measure-default-filters.png' class='rounded-gif' /> <br/> 

For detailed YAML configurations, see the [`filters`](/reference/project-files/canvas-dashboards#defaults) section in our reference documentation.-->

## Example Canvas Dashboards

Here are a few deployed examples of Canvas dashboards that you can check out!

- **[E-commerce demo dashboard](https://ui.rilldata.com/demo/ezcommerce-demo/canvas/canvas)**
- **[Programmatic advertising demo dashboard](https://ui.rilldata.com/demo/rill-openrtb-prog-ads/canvas/executive_overview)**
- **[New York City demo dashboard](https://ui.rilldata.com/demo/nyc-canvas-jam/canvas/scorecard%20canvas)**
- **[NYC party demo dashboard 🎉](https://ui.rilldata.com/demo/nyc-canvas-jam/canvas/Leaderboard)**

