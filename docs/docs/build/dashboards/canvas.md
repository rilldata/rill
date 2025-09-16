---
title: Canvas Dashboards
description: Create custom dashboards by assembling visualizations and components
sidebar_label: Canvas Dashboards
sidebar_position: 05
---

While Rill's signature **[Explore dashboards](/build/dashboards/dashboards.md)** let you slice, dice, and interact with your data in our suggested layout, **Canvas dashboards** let you define your own visualizations and arrange them into your preferred layout.


Canvas dashboards are built using various components that can display data, create visualizations, and add rich content from independent metrics views. You can create components dynamically through the visual Canvas Dashboard interface or define them in individual YAML files.

## Available Components

Canvas dashboards support three main types of components:

- **[Data Components](/build/dashboards/canvas-components/data)** - Display structured data and key metrics (KPIs, leaderboards, tables)
- **[Chart Components](/build/dashboards/canvas-components/chart)** - Create rich visualizations (bar charts, line charts, heat maps, etc.)
- **[Miscellaneous Components](/build/dashboards/canvas-components/misc)** - Add text, images, and other content

Each component page shows both the visual result and the corresponding YAML code, making it easy to understand how to implement them in your own dashboards.

For a complete overview of all available components, see our [**Canvas Components**](/build/dashboards/canvas-components) reference.

## Creating a Canvas Dashboard
### A Visual Editing Experience 

To modify any single widget, click to select it and use the righthand panel to change its associated properties. Click outside the widget to view global properties associated with the overall Canvas.

<img src = '/img/build/canvas/selected-widget.png' class='rounded-gif' />
<br/>


### Local Filters
Optionally toggle on the filter bar under Canvas properties to give dashboard viewers access to the same time and dimension filters available on Explore dashboards.

<img src = '/img/build/canvas/global-filter-bar.png' class='rounded-gif' />
<br/>

Filters for a single KPI, Chart, or Table can be separated from the global filters via the "Filters" tab in the properties panel, where you can set filters that are local to just that widget.


<img src = '/img/build/canvas/local-filters.png' class='rounded-gif' />


### Making changes to the YAML 
While we encourage creating Canvas dashboards via the visual editing experience described above, you can always edit the YAML file directly using the code view by toggling the switch next to the filename at the top of the page. Please see our [customization page](/build/dashboards/customization) and [reference documentation](/reference/project-files/canvas-dashboards) for more information.


<img src = '/img/build/canvas/code-toggle.png' class='rounded-gif' />


## Example Canvas Dashboards 
Here are a few deployed examples of Canvas dashboards that you can check out!

- **[E-commerce demo dashboard](https://ui.rilldata.com/demo/ezcommerce-demo/canvas/canvas)**
- **[Programmatic advertising demo dashboard](https://ui.rilldata.com/demo/rill-openrtb-prog-ads/canvas/executive_overview)**
- **[New York City demo dashboard](https://ui.rilldata.com/demo/nyc-canvas-jam/canvas/scorecard%20canvas)**
- **[NYC party demo dashboard 🎉](https://ui.rilldata.com/demo/nyc-canvas-jam/canvas/Leaderboard)**

