---
title: "Time Dimensions"
description: "Add multiple time dimensions to enable time-based filtering and analysis across different temporal columns"
sidebar_label: "Time Dimensions"
sidebar_position: 52
---

While every metrics view has a primary `timeseries` column that powers the main time series chart, you can define additional time dimensions using `type: time`. This enables users to filter and analyze data across multiple temporal columns.

## When to Use Time Dimensions

Time dimensions are useful when your data contains multiple date or timestamp columns that users may want to filter by:

- **Order date vs. Ship date**: Filter orders by when they were placed or when they were shipped
- **Created vs. Updated timestamps**: Analyze records by creation date or last modification
- **Event time vs. Processing time**: Distinguish between when events occurred and when they were recorded
- **Multiple business dates**: Handle scenarios like invoice date, due date, and payment date

## Adding a Time Dimension

To create a time dimension, add `type: time` to your dimension definition:

```yaml
version: 1
type: metrics_view

model: orders_model
timeseries: order_date  # Primary time series for charts

dimensions:
  - name: order_date
    display_name: Order Date
    column: order_date
    type: time

  - name: ship_date
    display_name: Ship Date
    column: ship_date
    type: time

  - name: customer_region
    column: region
```

## Time Dimension vs. Timeseries

Understanding the difference between these two concepts is important:

| Feature | `timeseries` (top-level) | `type: time` (dimension) |
|---------|--------------------------|--------------------------|
| Purpose | Powers main time series chart | Enables time-based filtering |
| Chart display | Shows trends over time | Used in filter panel |
| Required | No (but recommended) | No |
| Multiple allowed | No (one per metrics view) | Yes |

The primary `timeseries` column determines which dates appear on the x-axis of your time series visualizations. Time dimensions provide additional temporal filtering options in the dashboard filter panel.

## Full Example

Here's a complete example with multiple time dimensions:

```yaml
version: 1
type: metrics_view

model: sales_model
timeseries: transaction_date

dimensions:
  # Time dimensions
  - name: transaction_date
    display_name: Transaction Date
    column: transaction_date
    type: time

  - name: fulfillment_date
    display_name: Fulfillment Date
    column: fulfillment_date
    type: time

  - name: return_date
    display_name: Return Date
    column: return_date
    type: time

  # Categorical dimensions
  - name: product_category
    display_name: Product Category
    column: category

  - name: store_location
    display_name: Store Location
    column: store_id

measures:
  - name: total_sales
    display_name: Total Sales
    expression: SUM(amount)

  - name: order_count
    display_name: Order Count
    expression: COUNT(*)
```

## Using Expressions

You can also create time dimensions using expressions to transform or derive time values:

```yaml
dimensions:
  - name: order_month
    display_name: Order Month
    expression: DATE_TRUNC('month', order_date)
    type: time

  - name: fiscal_quarter_start
    display_name: Fiscal Quarter Start
    expression: DATE_TRUNC('quarter', order_date + INTERVAL '3 months') - INTERVAL '3 months'
    type: time
```

:::tip
Time dimensions work with columns of type `TIMESTAMP`, `TIME`, or `DATE`. If your source data stores dates in a different format (like strings), use an expression to convert them to a proper date type.
:::
