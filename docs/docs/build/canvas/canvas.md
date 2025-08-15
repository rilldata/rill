---
title: Create Canvas Dashboards
description: Create dashboards by assembling visualizations of various metrics views
sidebar_label: Create Canvas Dashboards
sidebar_position: 05
---

<div style={{ 
  position: "relative", 
  width: "100%", 
  paddingTop: "56.25%", 
  borderRadius: "15px",  /* Softer corners */
  boxShadow: "0px 4px 15px rgba(0, 0, 0, 0.2)"  /* Shadow effect */
}}>
  <iframe credentialless="true"
    src="https://www.youtube.com/embed/Idr2bvQw9_I?si=-xB0ppGOIavlmsE-"
    frameBorder="0"
    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
    allowFullScreen
    style={{
      position: "absolute",
      top: 0,
      left: 0,
      width: "100%",
      height: "100%",
      borderRadius: "10px", 
    }}
  ></iframe>
</div>
<br/>
----
While Rill's signature **[Explore dashboards](/build/dashboards/dashboards.md)** let you slice, dice, and interact with your data in our suggested layout, **Canvas dashboards** let you define your own visualizations of data from multiple [metrics views](/build/metrics-view/metrics-view.md) and arrange them into your preferred layout. 

You can add the following widgets to a Canvas dashboard:

- **Chart**: Select dimensions and measures to visualize as a line, bar, or donut chart.
- **KPI**: Create a grid of key performance indicators from measures defined in a metrics view.
- **Leaderboard**: Just like in our Explore dashboards, display a grid of your top lists per category and click to quickly filter.
- **Table**: Select dimensions and measures to visualize as a flat or nested table.
- **Text**: Use markdown to create labels or descriptive text blocks.
- **Image**: Provide a URL to an image.


## A Visual Editing Experience 

To modify any single widget, click to select it and use the righthand panel to change its associated properties. Click outside the widget to view global properties associated with the overall Canvas.

<img src = '/img/build/canvas/selected-widget.png' class='rounded-gif' />
<br/>


## Global and Local Filters
Optionally toggle on the filter bar under Canvas properties to give dashboard viewers access to the same time and dimension filters available on Explore dashboards.

<img src = '/img/build/canvas/global-filter-bar.png' class='rounded-gif' />
<br/>

Filters for a single KPI, Chart, or Table can be separated from the global filters via the "Filters" tab in the properties panel, where you can set filters that are local to just that widget.


<img src = '/img/build/canvas/local-filters.png' class='rounded-gif' />


## Making changes to the YAML 
While we encourage creating Canvas dashboards via the visual editing experience described above, you can always edit the YAML file directly using the code view by toggling the switch next to the filename at the top of the page. Please see our [customization page](customization.md) and [reference documentation](/reference/project-files/canvas-dashboards) for more information.


<img src = '/img/build/canvas/code-toggle.png' class='rounded-gif' />


## Define Dashboard Access

Along with [metrics views security policies](/build/metrics-view/security), you can set access on the dashboard level.  Access on dashboards will override the access rules in metrics views. This might be useful if there are a few dashboards that you want to limit to a set of users with a specific set of dimensions. 

Or, the dashboard needs some QA by the team with [full data](/connect/templating) before sharing to the rest of the team.

```yaml
security:
  access: '{{ has "partners" .user.groups }}' #only dashboard access can be defined here, other security policies must be set on the metrics view
```

## Example Canvas Dashboards 
Here are a few deployed examples of Canvas dashboards that you can check out!

- **[E-commerce demo dashboard](https://ui.rilldata.com/demo/ezcommerce-demo/canvas/canvas)**
- **[Programmatic advertising demo dashboard](https://ui.rilldata.com/demo/rill-openrtb-prog-ads/canvas/executive_overview)**
- **[New York City demo dashboard](https://ui.rilldata.com/demo/nyc-canvas-jam/canvas/scorecard%20canvas)**
- **[NYC party demo dashboard 🎉](https://ui.rilldata.com/demo/nyc-canvas-jam/canvas/Leaderboard)**

