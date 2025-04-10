---
title: Create Canvas Dashboards
description: Create dashboards by assembling visualizations of various metrics views
sidebar_label: Create Canvas Dashboards
sidebar_position: 05
---

<img src = '/img/build/canvas/RTBAds-Canvas-2.png' class='rounded-gif' />
<br/>


While Rill's signature **[Explore dashboards](/build/dashboards/dashboards.md)** let you slice-and-dice and interact with your data in our suggested layout, **Canvas dashboards** let you define your own visualizations of data from from multiple [metrics views](/build/metrics-view/metrics-view.md) and arrange them into your preferred layout. 

You can add the following widgets to a Canvas dashboard:

- **KPI**: Create a grid of key performance indicators from measures defined in a metrics view.
- **Chart**: Select dimensions and measures to visualize as a line or bar chart, optionally split by category.
- **Table**: Select dimensions and measures to visualize as a flat or nested table.
- **Text**: Use markdown to create labels or descriptive text blocks.
- **Image**: Provide a URL to an image.


## A Visual Editing Experience 

To modify any single widget, click to select it and use the righthand panel to change its associated properties. Click out of the widget to view global properties associated with the overall Canvas.

<img src = '/img/build/canvas/selected-widget.png' class='rounded-gif' />
<br/>


## Global and Local Filters
Optionally toggle on the filter bar under Canvas properties to give dashboard viewers access to the same time and dimension filters available on Explore dashboards.

<img src = '/img/build/canvas/global-filter-bar.png' class='rounded-gif' />
<br/>

Filters for a single KPI, Chart, or Table can be separated from the global filters via the "Filters" tab in the properties panel, where you can set filters local to just that widget.


<img src = '/img/build/canvas/local-filters.png' class='rounded-gif' />


## Making changes to the YAML 
While we encourage creating Canvas dashboards via the visual editing experience described above, you can always change the YAML file directly using the code view, by flipping the toggle next to the filename at the top of the page. Please see our [customimzation page](customization.md) and [reference documentation](/reference/project-files/canvas-dashboards) for more information.


<img src = '/img/build/canvas/code-toggle.png' class='rounded-gif' />


## Example Canvas Dashboards 
Here are a few deployed examples of Canvas dashboards that you can check out!

- **[E-commerce demo dashboard](https://ui.rilldata.com/demo/ezcommerce-demo/canvas/canvas)**
- **[Programmatic advertising demo dashboard](https://ui.rilldata.com/demo/rill-openrtb-prog-ads/canvas/executive_overview)**
- **[New York City demo dashboard](https://ui.rilldata.com/demo/nyc-canvas-jam/canvas/scorecard%20canvas)**
- **[NYC party demo dashboard 🎉](https://ui.rilldata.com/demo/nyc-canvas-jam/canvas/Leaderboard)**

