---
title: "Filters & Comparisons"
description: Filters & Comparisons
sidebar_label: "Filters & Comparisons"
sidebar_position: 10
---

Time is one of the most powerful dimensions in analytics, and Rill dashboards are designed to help you make the most of your time series data. The time filter is a central tool for exploring trends, identifying anomalies, and comparing performance across different periods. Whether you’re analyzing daily sales, monitoring system metrics, or tracking campaign results, understanding how to use time filters will help you unlock deeper insights and make more informed decisions.


<div style={{ 
  position: "relative", 
  width: "100%", 
  paddingTop: "56.25%", 
  borderRadius: "15px",  /* Softer corners */
  boxShadow: "0px 4px 15px rgba(0, 0, 0, 0.2)"  /* Shadow effect */
}}>
  <iframe credentialless="true"
    src="https://www.youtube.com/embed/1gmEBf2cv9U?si=bD2gXKAfW3Zb3FAn"
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



<img src = '/img/explore/filters/time-filter.png' class='rounded-gif' />
<br />

## Time Filters

Let's take a look of the different types of time filters that are available in Rill.

### Comparison Time Ranges

Comparing performance across different periods is a key part of time series analysis. Rill makes this easy with **comparison time ranges**:

- **Enable comparison:** Turn on comparison mode to see how your current period stacks up against a previous one (e.g., this week vs. last week, this month vs. last month).
- **Visual cues:** Rill highlights changes, trends, and deltas, making it easy to spot improvements or areas needing attention.
  
<img src = '/img/explore/filters/kpi_compare.png' class='rounded-gif' />
<br />

:::tip
Rill provides different options for time period comparison - by time period or by selected hours. 

For the former, you can let data "fill in" by selecting time period options like last day, previous 7 days, last week. Future periods will show 'no data.' Use cases here would be for pacing reports or seeing data refresh during business hours. 

For the latter, you can compare the full period looking with options like last 24 hours vs. prior 24 hours. In this case, the time series will be fully complete, comparing up to the most recent period vs. the same hour/day/week in prior periods.  
:::

### Filter by Scrubbing 

For a specific view into your time series graph, you can interactively scrub directly on the time series graph. 


<img src = '/img/explore/filters/scrub-graph.gif' class='rounded-gif' />
<br />

This allows the ability for a more detailed view into your time series without having to change the overall time series filter for quick access to measures. 



:::tip _as of latest_

At the right of the time filter, you’ll see _as of lastest_. This indicates the last time the data has been reloaded. If you're noticing the "Past 7 Days" is not indicating the correct dates, take a look to see when the last time the data successfully reloaded. When hovering over the text, you'll see the following.

- _**Earliest:**_ The first available timestamp in your dataset.
- _**Latest:**_ The most recent timestamp in your dataset (should match “as of”).
- _**Now:**_ The current system time.

**Why it matters:**  
This helps you understand how fresh your data is, when it was last updated, and whether you’re looking at real-time or historical information.

**Example:**
- *Earliest:* Mon, Jun 30, 2025, 6:00 PM  
- *Latest:* Wed, Jul 16, 2025, 5:00 PM  
- *Now:* Wed, Jul 16, 2025, 8:44 PM

:::
<!-- 
## Default Time Ranges

Rill provides a set of default time ranges that appear in the time filter dropdown. These are designed to cover the most common analysis periods, such as:

- **Today, Yesterday**
- **Last 7 days, Last 30 days**
- **Past 3 Months**
- **Year to date**

You can customize these ranges in your `explore` settings or `rill.yaml` file to better fit your organization’s needs. For example, you might add a “Last 90 days”, `P90D`, or remove less relevant ones.

```yaml
time_ranges:
  - PT6H
  - PT24H
  - P7D
  - P14D
  - P4W
  - P3M
  - P12M
  - rill-TD
  - rill-WTD
  - rill-MTD
  - rill-QTD
  - rill-YTD
  - rill-PDC
  - rill-PWC
  - rill-PMC
  - rill-PQC
  - rill-PYC
```

Each time range is defined using [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) syntax, or you can use [Rill’s ISO 8601 extensions](../../reference/rill-iso-extensions.md#extensions) for more flexibility. This allows you to precisely control the periods available to users.

---

## Time Zones

Time zones play a crucial role in time series analysis, especially for global teams or data collected across multiple regions. In Rill, you can configure the default time zones in your `explore` settings or `rill.yaml`, ensuring that dashboards reflect the correct local time for your users.

- **Default time zones:** Set a default for your organization or project.
- **User selection:** Users can override the default and select any time zone from a searchable list, making it easy to view data in their preferred context.

```yaml
time_zones:
  - UTC
  - America/Los_Angeles
  - America/Chicago
  - America/New_York
  - Europe/London
  - Europe/Paris
  - Asia/Jerusalem
  - Europe/Moscow
  - Asia/Kolkata
  - Asia/Shanghai
  - Asia/Tokyo
  - Australia/Sydney
```

:::tip

Users reporting different numbers or claiming the data is wrong? Double check that both users are looking at the same time zone! Rill saves a user's last state and if they were viewing a different time zone at some point, it might be saved to that time zone.

::: -->


### Time Grain

The **time grain** determines the level of detail shown in your time series charts and tables—such as hour, day, week, or month. This is controlled by the `smallest_time_grain` setting in your metrics view configuration.

- **Fine grain (e.g., hour):** Useful for short time ranges or when you need to see detailed patterns, such as hourly website traffic or system metrics.
- **Coarse grain (e.g., week, month):** Better for long-term trends or when you want to smooth out short-term fluctuations.

Rill automatically adjusts the available grains based on the selected time range. For example, if you select a year-long period, hourly data may be hidden to keep charts readable and performant.


If videos are more your jam, take a look at [our series of YouTube videos](https://www.youtube.com/watch?v=wTP46eOzoCk&list=PL_ZoDsg2yFKgi7ud_fOOD33AH8ONWQS7I&index=1) to get started!

## Dimension / Measure Filter

Rill is particularly suited for exploratory analysis - to be able to slice & dice data quickly. As such, there are a variety of filter types and filter mechanisms throughout the app. The goal for each Rill dashboard is to provide users with all metrics and dimensions required for each use case and create an interactive experience to cut data in any form.
<div style={{ 
  position: "relative", 
  width: "100%", 
  paddingTop: "56.25%", 
  borderRadius: "15px",  /* Softer corners */
  boxShadow: "0px 4px 15px rgba(0, 0, 0, 0.2)"  /* Shadow effect */
}}>
  <iframe credentialless="true"
    src="https://www.youtube.com/embed/LNyvn8lRFUw?si=FyEViEreF4cIrE09"
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

:::tip Human readable URL
With the release of v0.52, we have introduced an easy way to craft specific views by modifying the URL directly. As you select filters, time ranges, and default dimension and measures, the URL will change accordingly. 

.../explore/explore_dashboard?tr=P3M&compare_tr=rill-PP&f=dimension in ('exampleA', 'exampleB')
:::

### Add / Hide Dimensions and Metrics

Users can add or hide dimensions and metrics to a subset of fields they wish to see at any given time. At the top left, above the time series and above the top left leaderboard, you'll find the Measures & Dimensions selectors to add or hide from the page. In the example below, `network` and `country` are deselected so would be hidden from view.

<img src = '/img/explore/filters/hide.png' class='rounded-gif' />
<br />


:::tip hiding metrics and dimensions by default
You can also change settings in the dashboard configuration to hide certain fields by default. You may want to do this to make dashboards easier to use (less complicated, narrowed to most commonly used) and to improve performance (hide high cardinality dimensions or complicated expressions in metrics). For more details, check out [dashboard customizations](/build/dashboards/customize.md#setting-default-views-for-dashboards).

Or, an administrator can set the default view of a dashboard by [bookmarking the view](../bookmarks.md) as Home. 
:::

### Filter by Dimensions

The primary/easiest way to filter data is by selecting values in the dimension tables. Leaderboards within Rill are fully interactive. Selecting any dimension in the table will automatically filter the remaining leaderboards and metrics by that selection. 

To add or remove dimensions on the page - select the All Dimensions picker above the Leaderboards. Next to the All Dimensions picker, you can also change which Metric is being highlighted to be able to update the entire page and cycle through each dimension table sorted by each metric.

You can also expand each dimension table to see all metrics and full list of those dimensions. In the expanded Leaderboard, you can search for dimension values, select all values returned, or exclude values from the result. 

Any filter applied in the Leaderboard will also show up in the filter bar at next to the time picker. You can apply the same search capabilities and select features in the filter bar as well.

<img src = '/img/explore/filters/filter.gif' class='rounded-gif' />
<br />


### Filter by Metrics

There are also use cases where you want to filter by the metric values returned. As an example - all customers over $1000 in revenue, all campaigns with at least 1 million impressions, all delivery locations with late times over 4%, etc. 

To add or remove metrics on the page - select the All Measures picker above the Time Series charts. 

These metric filters can be applied from the filter bar. To apply a metric filter:

- Select the metric you wish to filter by (e.g. Total Cost)
- Select which dimension to sort/key the metric by (e.g. Cost by Region)
- Select your Threshold Type (e.g. Great Than)
- Input your Threshold amount and Click Enter


<img src = '/img/explore/filters/image.png' class='centered' />
<br />


:::tip
Metric filters are a good way to "sort" by two different metrics. First, apply your metric threshold. Then, sort by your metrics within the Leaderboard to do multi-metric sorting. 

As an example - to see most active enterprise customers - filter all customers with revenue greater than $1000 then sorted by number of users increased descending.
:::


### Dimension Comparisons
In addition to time comparisons, you can select multiple dimension values to compare trends of those specific data points. Select the comparison option on the top left of any leaderboard and select multiple dimensions

Deselect the comparison option or clear the filter bar to remove your comparisons.

<img src = '/img/explore/filters/comparison.gif' class='rounded-gif' />
<br />
:::note
For more advanced time and dimension comparisons, visit the [Time Dimension Detail](/explore/dashboard-101/tdd) page.
:::

