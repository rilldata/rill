---
title: "👋 Explore Quickstart"
slug: "getting-started"
excerpt: "Analysis basics and capabilities overview"
hidden: false
createdAt: "2021-06-17T00:27:50.336Z"
updatedAt: "2022-07-13T07:15:20.929Z"
---
# Explore Overview"
}
[/block]
Any metric or dimension in your Druid dataset is available for your Explore dashboards. In Explore, you can create a data source that divides those fields into any subset, transformation (via calculations, lookups, formatting), or re-ordering depending on your dashboard needs.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/0414458-Explore_Overview.png",
        "Explore Overview.png",
        2002,
        539,
        "#ebeff1"
      ]
    }
  ]
}
[/block]

[block:callout]
{
  "type": "info",
  "body": "By default, Rill creates a Staging dashboard which includes all Dimensions and Metrics in your data source which is then editable or can be leveraged to create customized views.",
  "title": "Default Dashboards"
}
[/block]

# Basic Analysis & Navigation"
}
[/block]
### Metrics & Dimensions

Metrics are any numerical value you have provided in your data (e.g. sales, impressions, eCPM, etc.). Dimensions are the lists of fields categorizing that data (e.g. DMA, publisher, account owner, etc.). 

Dashboards provide the ability to:

  * Add/remove metrics and dimensions
  * Expand dimensions to see full list of data
  * Download results across metrics and dimensions

### Visualizations
Metrics and dimensions can be visualized in a variety of ways including:

  * topN tables 
  * Time series
  * Bar charts
  * Multi-line/bar graphs
  * Pivot tables
  
### Filtering
Throughout the Dashboard, you are able to filter data including:

  * Include dimension values
  * Exclude dimension values
  * Intersections of multiple dimensions (pivot tables) 
  * Default time periods (e.g. last day, week, 30 days)
  * Custom time periods (e.g. 07/12 thru 07/18, Tuesdays only)
  * Adjusting timezones to troubleshoot global customer questions

To filter dimension values, select the + next to the dashboard title or filter within a dimension table. You can search the list of dimension values to find a specific value if it does not appear in the list.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/de9288f-Dim_Filter.png",
        "Dim Filter.png",
        2685,
        728,
        "#f8f9f9"
      ],
      "sizing": "80",
      "caption": "Dimension Filter Bar"
    }
  ]
}
[/block]

[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/8e50202-Dim_Filter_Search.png",
        "Dim Filter Search.png",
        2685,
        728,
        "#f9fafa"
      ],
      "sizing": "80",
      "caption": "Search for values within a Dimension"
    }
  ]
}
[/block]
### Search
Explore's search bar provides the ability to search all data for potential results and filtering. Search is useful for identifying all dimensions that contain a specific value or for quick filtering of a value in all places. 

Search is the magnifying glass icon located in the top right.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/9299736-Search_Bar.png",
        "Search Bar.png",
        2685,
        563,
        "#abccd8"
      ],
      "sizing": "80"
    }
  ]
}
[/block]
Search results cover all dimensions and return values across all dimensions or selected dimensions.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/7ab968d-Search_Results.png",
        "Search Results.png",
        2685,
        944,
        "#eff1f2"
      ],
      "sizing": "80"
    }
  ]
}
[/block]
### Chart Comparisons
Dashboards provide the ability to compare dimensions to each other and across time to identify outliers and insights:

  * Compare metrics to prior periods - either summary or by time series
  * Compare dimension trend lines (select multiple dimensions and compare results)
  * Compare share of dimensions (multi-bar charts)

# Example Navigation Video

The example video below includes basic time and dimension filtering along with navigating comparison charts and drilling down into data further.

```
<div style=\"position: relative; padding-bottom: 44.063647490820074%; height: 0;\"><iframe src=\"https://www.loom.com/embed/c4b44a2d5b7c42128a9fcabb100f25da\" frameborder=\"0\" webkitallowfullscreen mozallowfullscreen allowfullscreen style=\"position: absolute; top: 0; left: 0; width: 100%; height: 100%;\"></iframe></div>
```

# Capabilities Overview

Explore provides a variety of capabilities to improve your time to insight. 

The list below is a subset of the most common features for improved analysis and reporting:

  * [Set Alerts](https://enterprise.rilldata.com/docs/alerting)
  * [Download Data & Schedule Exports ](https://enterprise.rilldata.com/docs/scheduled-exports)
  * [Set Bookmarks (saved views of your dashboards)](https://enterprise.rilldata.com//docs/bookmarking)
  * [Create pivot tables for analysis or scheduled reports using Facet](https://enterprise.rilldata.com/docs/facet-pivot-table-splits)