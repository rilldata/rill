---
title: Canvas Dashboard
sidebar_label: "Canvas Dashboard"
sidebar_position: 05
---


After logging into [Rill Cloud](https://ui.rilldata.com), you should see all projects within your [organization](/manage/organization-management#organization) that is available and/or has been granted permissions to your user profile. Within each project, you'll then be able to access the corresponding individual dashboards that belong to a particular Rill project. 


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

<img src = '/img/explore/canvas/canvas-dashboard.png' class='rounded-gif' />
<br />


Similar to our [Explore dashboards](/explore/dashboard-101), Canvas Dashboards also include a similar navigation bar to control the dashboard components.

### Navigation Bar

- _**Time Selector and Time Selector Comparison:**_ You can change the period of analysis to different ranges of time (see `red` box), either by selecting from a pre-defined period (such as last week) or choosing a custom date range. Along with this, you can enable a comparison filter to compare range of dates with 1 click.

- _**Filtering:**_ Underneath the time selector, you'll also be able to find your filter bar (see `orange` box) where you can [add filters](/explore/filters/filters.md) for metrics (e.g. `campaigns>1000`) or for dimensions (e.g. `campaign_name = Instacart`).

:::tip identical names in metrics views

 If your dimensions or measures have the same name in your metrics view, filters will apply to all components, regardless if it's in a different metrics view.
 :::

<!-- - _**Alerts, Bookmarks and Sharing:**_ You can create an [alert](/explore/alerts) by selecting the bell, customizing the default view of the dashboard (see `purple` box) to a predefined set of metrics, dimensions, and filters by selecting the [bookmark](/explore/bookmarks.md), or share the dashboard ([internally by clicking the `Share` button](/manage/user-management#admin-invites-user-from-rill-cloud) or [externally via Public URLs](/explore/public-url.md)) . -->

## Component Navigation
<img src = '/img/explore/canvas/canvas-navigaton.png' class='rounded-gif' />
<br />


If you want to further drill into a component's data, select the top right button to take you to the equivalent explore dashboard.

:::tip no button?

If no explore dashboard exists, and/or you don't have [permissions to view it](/build/dashboards/#define-dashboard-access), no button will appear and is as designed.

:::