---
title: "Dashboard Quickstart"
description: Dashboard Quickstart
sidebar_label: "Dashboard Quickstart"
sidebar_position: 15
---

<div style={{ 
  position: "relative", 
  width: "100%", 
  paddingTop: "56.25%", 
  borderRadius: "15px",  /* Softer corners */
  boxShadow: "0px 4px 15px rgba(0, 0, 0, 0.2)"  /* Shadow effect */
}}>
  <iframe credentialless="true"
    src="https://www.youtube.com/embed/wTP46eOzoCk?si=9JzY-CuzqQU4uMiR"
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


## Overview

After logging into [Rill Cloud](https://ui.rilldata.com), you should see all projects within your [organization](/manage/organization-management#organization) that is available and/or has been granted permissions to your user profile. Within each project, you'll then be able to access the corresponding individual dashboards that belong to a particular Rill project. 

<img src = '/img/explore/dashboard101/rill-cloud-landing-page.png' class='rounded-gif' />
<br />


## Navigating the Dashboard

<img src = '/img/explore/dashboard101/quickstart.png' class='rounded-gif' />
<br />


**Explore** 
The main screen of any Rill dashboard is called the _Explore_ page. As seen above, this is divided into three section. 

- Navigation Bar
- Measures panel (Left)
- Dimensions Leaderboard (Right)

### Navigation Bar

- _**Time Selector and Time Selector Comparison:**_ You can change the period of analysis to different ranges of time (see `red` box), either by selecting from a pre-defined period (such as last week) or choosing a custom date range. Along with this, you can enable a comparison filter to compare range of dates with 1 click.

- _**Filtering:**_ Underneath the time selector, you'll also be able to find your filter bar (see `orange` box) where you can [add filters](/explore/filters) for metrics (e.g. `campaigns>1000`) or for dimensions (e.g. `campaign_name = Instacart`).

- _**Explore or Pivot:**_ You can switch the view from _explore_ to [_pivot_](/explore/dashboard-101/pivot) by selecting either from the UI (see `pink` box)

- _**Alerts, Bookmarks and Sharing:**_ You can create an [alert](/explore/alerts) by selecting the bell, customizing the default view of the dashboard (see `purple` box) to a predefined set of metrics, dimensions, and filters by selecting the [bookmark](/explore/bookmarks), or share the dashboard ([internally by clicking the `Share` button](/manage/user-management#admin-invites-user-from-rill-cloud) or [externally via Public URLs](/explore/public-url)) .


### KPI Widget (Measures) Panel
<div style={{ 
  position: "relative", 
  width: "100%", 
  paddingTop: "56.25%", 
  borderRadius: "15px",  /* Softer corners */
  boxShadow: "0px 4px 15px rgba(0, 0, 0, 0.2)"  /* Shadow effect */
}}>
  <iframe credentialless="true"
    src="https://www.youtube.com/embed/Dqkfp6F_9y4?si=z-22kqFd5dhQA6w8"
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



- _**Measures:**_  All _**metrics**_ that are available in the underlying model \ are viewable on the left-hand side, broken out with summary numbers (e.g. eCPM) and timeseries visualizations (based on your configured `timeseries` column in your [dashboard YAML](/reference/project-files/explore-dashboards)). You can add, remove or reorder your metrics from the page by simply selecting them from the dropdown above the charts (see `yellow` box). If you select any specific measure, you will be navigating to the [Time Dimension Detail](/explore/dashboard-101/tdd).

- _**Time Dimension Detail:**_ A detailed view of a single specific measure that can be further drilled down to understand minute details in your data. As with the Explore page, you can add comparison dimensions to visualize the value for several specific dimension values. For more information see: [Time Dimension Detail](/explore/dashboard-101/tdd).

:::note Big Number Formatting

[Formatting of your measures](/build/metrics-view#measures) will not change the granularity of the Big Number, but you'll see the formatting being applied to the TDD, Dimension Leaderboard, and Pivot tables. 
:::

### Dimensions Leaderboard Panel

<div style={{ 
  position: "relative", 
  width: "100%", 
  paddingTop: "56.25%", 
  borderRadius: "15px",  /* Softer corners */
  boxShadow: "0px 4px 15px rgba(0, 0, 0, 0.2)"  /* Shadow effect */
}}>
  <iframe credentialless="true"
    src="https://www.youtube.com/embed/aQQBFHbLrMQ?si=il-w_ssQmGrqCfsO"
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


- _**Dimensions:**_  All _**dimensions**_ available in the underlying model on the right-hand side via leaderboard / toplist charts. You can add, remove or re-order any dimension from the page by simply selecting them from the dropdown above the charts (see `green` boxes). You can also drill into leaderboards further (see `blue` box) to see all corresponding metrics for a specific dimension. Within that drilldown, you can also then sort by metric, search your dimensions, and/or [export data](/explore/exports). It is also possible to display [multiple measures in the dimension leaderboard](/explore/dashboard-101/multi-metrics).


:::info Search for individual attributes


After drilling into a leaderboard (or what we sometimes refer to as a _toplist_ chart), rather than scrolling and finding an individual attribute (especially if the list is very long), you can also quickly search for a value and select / apply it to your dashboard by using the upper-right search box.


<img src = '/img/explore/dashboard101/search-box.png' class='rounded-gif' />
<br />

:::




:::tip Don't have a Rill project or dashboard deployed yet?
If you want to get hands on and see what interacting with a Rill dashboard feels like, we have a set of [demo projects](https://ui.rilldata.com/demo) already deployed on Rill Cloud and publicly available for everyone to try out. These [same projects](/#examples) are also available on GitHub and can be deployed locally using Rill Developer.
:::


For more details about additional capabilities and/or how to utilize more advanced functionality within Rill dashboards, please see the [reference](#reference) section.


### Keyboard shortcuts
Whether you need to see the full value of a long JSON, or copy a value, there are some available keyboard shortcuts in the Rill Cloud Dashboards. More coming soon!

List of commands:
- __*Copy values*__ ( ``shift + click`` ) - Copy the value of the row value. 
- __*Value previewer*__ ( ``space`` ) - See the full text value of the row value.
- __*Lock Insepector*__ ( ``L`` ) - Lock the inspector (allows scrolling through long values)

<img src = '/img/explore/dashboard101/preview-value.png' class='rounded-gif' />
<br />


## Banners!
Another additional feature that you can add to an Explore dashboard are banners. Whether it is to inform your end-users about specific guidelines on how to use Rill, or an informational post about the datasets being used, you can design the banner to whatever text you'd like.

Simple add the following to your explore-dashboard.yaml 

```yaml
banner: Your custom message here!
```

<img src = '/img/explore/dashboard101/banner.png' class='rounded-gif' />
<br />

## Reference

- [Filters & Comparisons](/explore/filters)
- [Bookmarks & Sharing](/explore/bookmarks)
- [Exports & Scheduled Reports](/explore/exports)
- [Public URL](/explore/public-url)
- [Alerts](/explore/alerts)
