---
title: "Dashboard Quickstart"
description: Dashboard Quickstart
sidebar_label: "Dashboard Quickstart"
sidebar_position: 00
---


## Overview


Depending on whether you are opening Rill Developer or logging into Rill Cloud, you will either the default "Getting started" landing page or a list of all projects available to your user. For the purposes of this article, we will assume that your project has already been [deployed to Rill Cloud](../deploy/deploy-dashboard/) and that you are looking to consume your dashboards in a production capacity.


After logging into [Rill Cloud](https://ui.rilldata.com), you should see all projects within your [organization](/manage/project-management#organization) that is available and/or has been granted permissions to your user profile. Within each project, you'll then be able to access the corresponding individual dashboards that belong to a particular Rill project. 

![Rill Cloud landing page](../../static/img/explore/dashboard101/rill-cloud-landing-page.png)


## Navigating the Dashboard

![quickstart](../../static/img/explore/dashboard101/quickstart.png)

**Explore** 
The main screen of any Rill dashboard is called the _Explore_ page. As seen above, this is divided into three section. 

- Navigation Bar
- Metrics panel
- Dimensions leaderboard

### Navigation Bar

- _**Time Selector and Time Selector Comparison:**_ You can change the period of analysis to different ranges of time (see `red` box), either by selecting from a pre-defined period (such as last week) or choosing a custom date range. Along with this, you can enable a comparison filter to compare range of dates with 1 click.

- _**Filtering:**_ Underneath the time selector, you'll also be able to find your filter bar (see `orange` box) where you can [add filters](filters/filters.md) for metrics (e.g. `campaigns>1000`) or for dimensions (e.g. `campaign_name = Instacart`).

- _**Explore or Pivot:**_ You can switch the view from _explore_ to [_pivot_](https://docs.rilldata.com/explore/filters/pivot) by selecting either from the UI (see `pink` box)

- _**Alerts, Bookmarks and Sharing:**_ You can create an [alert](./alerts/alerts.md) by selecting the bell, customizing the default view of the dashboard (see `purple` box) to a predefined set of metrics, dimensions, and filters by selecting the [bookmark](bookmarks.md), or share the dashboard ([internally by clicking the `Share` button](/manage/user-management#admin-invites-user-from-rill-cloud) or [externally via Public URLs](./public-url.md)) .


### Metrics Panel

- _**Measures:**_  All _**metrics**_ that are available in the underlying model \ are viewable on the left-hand side, broken out with summary numbers (e.g. eCPM) and timeseries visualizations (based on your configured `timeseries` column in your [dashboard YAML](/reference/project-files/explore-dashboards.md)). You can add or remove any metric from the page by simply selecting them from the dropdown above the charts (see `yellow` box). If you select any specific measure, you will be navigating to the [Time Dimension Detail](https://docs.rilldata.com/explore/filters/tdd).

### Dimensions Leaderboard Panel

- _**Dimensions:**_  All _**dimensions**_ available in the underlying model on the right-hand side via leaderboard / toplist charts. You can add or remove any dimension from the page by simply selecting them from the dropdown above the charts (see `green` boxes). You can also drill into leaderboards further (see `blue` box) to see all corresponding metrics for a specific dimension. Within that drilldown, you can also then sort by metric, search your dimensions, and/or [export data](exports.md). 
:::info Search for individual attributes

After drilling into a leaderboard (or what we sometimes refer to as a _toplist_ chart), rather than scrolling and finding an individual attribute (especially if the list is very long), you can also quickly search for a value and select / apply it to your dashboard by using the upper-right search box.

![Using the search box within a leaderboard](../../static/img/explore/dashboard101/search-box.png)

:::




:::tip Don't have a Rill project or dashboard deployed yet?
If you want to get hands on and see what interacting with a Rill dashboard feels like, we have a set of [demo projects](https://ui.rilldata.com/demo) already deployed on Rill Cloud and publicly available for everyone to try out. These [same projects](/home/get-started#example-projects-repository) are also available on Github and can be deployed locally using Rill Developer.
:::


For more details about additional capabilities and/or how to utilize more advanced functionality within Rill dashboards, please see the [reference](#reference) section.

## Reference

- [Filters & Comparisons](filters/filters.md)
- [Bookmarks & Sharing](bookmarks.md)
- [Exports & Scheduled Reports](exports.md)
- [Public URL](public-url.md)
- [Alerts](/explore/alerts/alerts.md)
