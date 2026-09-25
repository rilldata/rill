---
title: Canvas Dashboards
sidebar_label: "Canvas Dashboards"
sidebar_position: 20
---


After logging into [Rill Cloud](https://ui.rilldata.com), you should see all projects within your [organization](/guide/administration/organization-settings#organization) that are available and/or have been granted permissions to your user profile. Within each project, you'll then be able to access the corresponding individual dashboards that belong to a particular Rill project.


<div style={{ 
  position: "relative", 
  width: "100%", 
  paddingTop: "56.25%", 
  borderRadius: "15px",  /* Softer corners */
  boxShadow: "0px 4px 15px rgba(0, 0, 0, 0.2)"  /* Shadow effect */
}}>
  <iframe credentialless="true"
    src="https://www.youtube.com/embed/z3ZHqypdGgc?si=X_oH9_wgNaiGzKOZ"
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

Prefer video? Check out our [YouTube playlist](https://www.youtube.com/watch?v=wTP46eOzoCk&list=PL_ZoDsg2yFKgi7ud_fOOD33AH8ONWQS7I&index=1) for a quick start!



## Navigating the Dashboard

![Canvas Dashboard](/img/explore/canvas/canvas-dashboard.png)


Similar to our [Explore dashboards](/guide/dashboards/explore), Canvas Dashboards also include a similar navigation bar to control the dashboard components.

### Navigation Bar

- _**Time Selector and Time Selector Comparison:**_ You can change the period of analysis to different ranges of time (see `red` box), either by selecting from a pre-defined period (such as last week) or choosing a custom date range. Along with this, you can enable a comparison filter to compare a range of dates with one click.

- _**Filtering:**_ Underneath the time selector, you'll also be able to find your filter bar (see `orange` box) where you can [add filters](/guide/dashboards/filters) for metrics (e.g. `campaigns>1000`) or for dimensions (e.g. `campaign_name = Instacart`).

:::tip identical names in metrics views

 If your dimensions or measures have the same name in your metrics view, filters will apply to all components, regardless if it's in a different metrics view.
 :::

<!-- - _**Alerts, Bookmarks and Sharing:**_ You can create an [alert](/guide/alerts) by selecting the bell, customizing the default view of the dashboard (see `purple` box) to a predefined set of metrics, dimensions, and filters by selecting the [bookmark](/guide/dashboards/bookmarks.md), or share the dashboard ([internally by clicking the `Share` button](/guide/administration/users-and-access/user-management#admin-invites-user-from-rill-cloud) or [externally via Public URLs](/guide/dashboards/public-urls.md)) . -->

## Component Navigation
![Canvas Navigation](/img/explore/canvas/canvas-navigaton.png)


If you want to further drill into a component's data, select the top right button to take you to the equivalent Explore dashboard.

:::tip no button?

If no Explore dashboard exists, and/or you don't have [permissions to view it](/developers/build/dashboards/customization#define-dashboard-access), no button will appear and is as designed.

:::

## Pin and Require Filters

Dashboard authors can keep specific filters in the filter bar, and can block a canvas until the viewer picks a value for them. This is useful when a canvas only makes sense for one region, customer, or account at a time.

| Setting | Effect for viewers |
|---------|--------------------|
| **Pinned** | The filter always appears in the filter bar, even when it has no value. Viewers can change its value but cannot remove the filter. |
| **Required** | The canvas does not load until the filter has a value. A required filter is always pinned. |

### Configure Pinned and Required Filters in YAML

List dimension or measure names under `filters` in the canvas file:

```yaml
type: canvas
display_name: "Regional bids overview"

filters:
  enable: true
  pinned:
    - device_type
  required:
    - auction_type
```

Because required filters are implicitly pinned, you do not need to list a name under both `pinned` and `required`.

Each name in `required` must be a dimension or measure on a metrics view used by the canvas's components. If it is not, the canvas fails to reconcile with an error like:

```
required filter "auction_typ" is not a dimension or measure on any metrics view referenced by this canvas
```

See the [`filters`](/reference/project-files/canvas-dashboards#filters) reference for the full list of properties.

### Configure Pinned and Required Filters in the Editor

When you edit a canvas in [Rill Developer](/developers/build/dashboards/canvas) or in Rill Cloud, each dimension and measure filter's dropdown shows two extra controls in its top-right corner:

- The **asterisk** marks the filter as required.
- The **pin** icon pins the filter.

![Pin and required controls in a canvas filter dropdown](/img/explore/canvas/pin-required-filter-toggles.png)

These controls change only your current editing session. To write them to the canvas file, click **Save as default** in the canvas editor header. This saves the pinned and required filters to `filters.pinned` and `filters.required`, and also saves the current filter values and time range as the canvas's [`defaults`](/reference/project-files/canvas-dashboards#defaults).

### What Viewers See

Until every required filter has a value, the canvas body is replaced by a **Select a value to continue** message that lists the missing filters, and the missing filters are highlighted in the filter bar. Once the viewer selects a value for each one, the canvas loads. Clearing a required filter's value blocks the canvas again.

![A canvas blocked by a required filter](/img/explore/canvas/required-filter-blocked.png)

:::tip Start viewers with a value
If you want the canvas to load immediately but never without a value, save a default value for the required filter. Viewers start with that value and can switch to another one, but cannot clear it without blocking the canvas.
:::

:::note Exports
A canvas with missing required filters has nothing to export. Set a value for every required filter before you export the canvas to PDF or schedule a PDF report of it.
:::
