---
title: Adhoc Measures in APIs
description: Define measures at query time with an arithmetic expression over existing measures
sidebar_label: Adhoc Measures
sidebar_position: 35
---

An adhoc measure is a measure you define in the query itself, by combining the metrics view's existing measures with an arithmetic expression such as `revenue - cost`. It is the API counterpart of the [adhoc measures](/guide/dashboards/explore/adhoc-measures) that users create in Explore dashboards. Use one when a client needs a derived value that isn't worth adding to the [metrics view](/developers/build/metrics-view).

An adhoc measure uses the `expression` compute:

| Field | Description |
|-------|-------------|
| `expression` | Required. An arithmetic expression over the metrics view's measure names. See the [expression reference](/guide/dashboards/explore/adhoc-measures#expression-reference). |
| `display_name` | Optional. The label used in exports and result metadata. |

The measure's `name` is the column name in the result. It must not match a dimension or measure in the metrics view.

## In a custom API

A custom API defined with a `metrics` query can include adhoc measures next to regular ones. This API returns revenue and profit margin by region:

```yaml
# apis/margin_by_region.yaml
type: api
metrics:
  metrics_view: sales_metrics
  dimensions:
    - name: region
  measures:
    - name: revenue
    - name: margin_pct
      compute:
        expression:
          expression: round((revenue - cost) / revenue * 100, 1)
          display_name: Margin %
  sort:
    - name: margin_pct
      desc: true
```

```json
[
  {"region": "US", "revenue": 100, "margin_pct": 40},
  {"region": "DK", "revenue": 50, "margin_pct": 10}
]
```

You can sort on an adhoc measure by its `name`, like any other measure. See [Calling APIs](/developers/build/custom-apis/calling) for how to call the endpoint.

## In the metrics view query APIs

The runtime's `MetricsViewAggregation` API accepts the same compute. In the REST API, the `expression` object goes directly on the measure:

```bash
curl -X POST "http://localhost:9009/v1/instances/default/queries/metrics-views/sales_metrics/aggregation" \
  -H "Content-Type: application/json" \
  -d '{
    "dimensions": [{"name": "region"}],
    "measures": [
      {"name": "revenue"},
      {"name": "margin_pct", "expression": {"expression": "(revenue - cost) / revenue", "displayName": "Margin %"}}
    ],
    "sort": [{"name": "margin_pct", "desc": true}]
  }'
```

The `MetricsViewTimeSeries` API takes adhoc measures in `ephemeralMeasures`, alongside or instead of `measureNames`. Only the `expression` compute is supported there:

```bash
curl -X POST "http://localhost:9009/v1/instances/default/queries/metrics-views/sales_metrics/timeseries" \
  -H "Content-Type: application/json" \
  -d '{
    "measureNames": ["revenue"],
    "ephemeralMeasures": [
      {"name": "avg_order_value", "expression": {"expression": "round(revenue / orders, 2)"}}
    ],
    "timeGranularity": "TIME_GRAIN_DAY"
  }'
```

## Validation and security

Rill parses the expression before it runs the query and returns an error when the expression is invalid, for example when it references an unknown measure or uses an aggregate function such as `sum()`. It never passes your expression text to the database: it generates the SQL from the parsed expression, and division by zero returns `null`.

Referenced measures are subject to the metrics view's [security policies](/developers/build/metrics-view/security). A caller can only reference measures they have access to, and row filters apply to the result as they do for any other query.
