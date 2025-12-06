---
title: "Pivot and Flat Tables"
description: Pivot
sidebar_label: "Pivot and Flat Tables"
sidebar_position: 30
---


Rill's multi-dimensional pivot provides lightning-fast, Excel-style pivot tables across multiple dimensions, time granularities, and metrics. The interface allows users to drag and drop any combination of fields for quick comparison and report generation. While the Explore view is great for data discovery, the Pivot view serves other use cases like:

- Regular reporting: Monthly financial reports, campaign reporting, client reports
- Time analysis: Dimensions by time columns for daily or weekly comparisons 
- Data export: Scheduled exports for generating CSVs for external analysis (_coming soon_)


<div style={{ 
  position: "relative", 
  width: "100%", 
  paddingTop: "56.25%", 
  borderRadius: "15px",  /* Softer corners */
  boxShadow: "0px 4px 15px rgba(0, 0, 0, 0.2)"  /* Shadow effect */
}}>
  <iframe credentialless="true"
    src="https://www.youtube.com/embed/mx6o2aAdlEc?si=PAKPa31r-k5pdW8l"
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


## Using Rill's Pivot View

Pivots can be split by time, measures, and dimensions. You can drag and drop any of these options to rows and columns to create your pivot table (and can drag between columns/rows as well). Dimensions are also searchable in the top left, or select the ```+``` to see the dimension list. As each dimension is added, you can track the kernels on the top left to see which rows have been added to the table. Each column dimension will also be visualized in the pivot. Like Time Dimension Detail, the same filter bar can be used to filter the pivot table overall. 

<img src = '/img/explore/pivot/pivot-overview.png' class='rounded-gif' />
<br />

:::note Measures as splits
Measures can only be used as column splits to define which metric is to be used in the table and cannot be added to the rows filter
:::

As new dimensions are added to the table, the comparison will be aggregated by default and can be expanded as you navigate the table. Pivot tables can then be scrolled through to quickly navigate the row/column combinations. If a given split does not have any valid data, ```no data``` will be displayed in the cell. Select the window button on the top left of the pivot table to expand the view to full screen. To quickly reset all the expanded rows in the pivot table, you can select collapse all in the top left. 


<img src = '/img/explore/pivot/pivot.gif' class='rounded-gif' />
<br />

## Flat Tables

While Pivot tables are a powerful tool to slice and dice your data into different configurations, sometimes a simple flat table is all you need. By selecting the toggle next to `Columns`, you can swap between pivot and flat tables. Unlike Pivot, which gives you options to add both row and column dimensions, a flat table will only allow you to add columns.

<img src = '/img/explore/pivot/flat-table.png' class='rounded-gif' />
<br />