---
title: "Edit Dashboard Metrics"
slug: "explore-json-examples"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt text="Illustrative snippets for updating your dashboards"/>

## Adding a Dashboard in Explore

To get started, a Staging dashboard will be created for your each datasource in your account. These staging dashboards will contain a full set of dimensions and metrics available in your datasource.

We recommend starting with a single dashboard per datasource that contains all metrics and dimensions. If there are common views of the data, [bookmarks are a helpful tool](/bookmarking) to save a specific set of filters, metrics and dimensions.

To add or edit dashboards, please contact your TAM or email [support@rilldata.com](mailto:support@rilldata.com) to make sure Admin capabilities are available for your account.

## Working with Dimensions

Each Dimension definition contains a short list of fields:
- a "bucket" (the type of dimension - 99% of cases will be "identity") 
- declared name (the name in the dashboard Datasource)
- an attribute (the name in the Druid database)
- a title (for both single and plural)
- a description (optional)

Each dimension (except for the last) is followed by a comma.

```json
 {
  "bucket": "identity",
  "name": "api_frameworks",
  "attribute": "api_frameworks",
  "titleSingle": "API Framework",
  "titlePlural": "API Frameworks",
  "description": "Framework used by API"
},
```
For dimensions with lookups, the JSON has a separate extraction function.

To add a lookup, a second statement is added including the type (registeredLookup), the lookup, and the ability to retain or remove missing values.

If the lookup is 1:1, set injective to true to improve performance.
```json
{
  "bucket": "identity",
  "name": "device_type",
  "attribute": "device_type",
  "titleSingle": "Device Type",
  "titlePlural": "Device Type",
  "extractionFn": {
    "type": "registeredLookup",
    "lookup": "example_lookup_table",
    "retainMissingValue": true,
    "injective": false,
    "optimize": true
    }
},
```
Beyond lookups to tables, lookup functions can also be used to replace values in a dimension with simple transform logic. 

Below the first, simpler example replaces missing values with "Not Available."

The second example replaces those Not Available values and also maps them to "False."
```json
{
  "bucket": "identity",
  "name": "ad_network_qtl",
  "attribute": "ad_network_id",
  "titleSingle": "Ad Network Name",
  "titlePlural": "Ad Network Name",
  "description": "Ad Network that owns the package",
  "extractionFn": {
    "retainMissingValue": false,
    "lookup": "ad_network_lookup",
    "replaceMissingValueWith": "Not Available",
    "type": "registeredLookup",
    "optimize": true,
    "injective": false
    }
},
 {
   "bucket": "identity",
   "name": "placement_flat_cpm_enabled",
   "attribute": "placement_flat_cpm_enabled",
   "titleSingle": "Flat CPM Placement",
   "titlePlural": "Flat CPM Placement",
   "extractionFn": {
      "type": "lookup",
      "dimension": "placement_flat_cpm_enabled",
      "outputName": "placement_flat_cpm_enabled_val",
      "replaceMissingValueWith": "False",
      "retainMissingValue": false,
      "lookup": {
        "type": "map",
        "map": {
          "True": "True",
           "False": "False",
           "Not Available": "False"
           }
       }
    }
}  
```

## Working with Metrics

Metrics passed directly within Druid are also straightforward and similar to basic dimensions by providing:
- name
- type of aggregate (sum, count, max, min, average)
- attribute (database field to aggregate)
- title (display name)
- description (*optional mouseover*)


```json
{
  "name": "bid_request_cnt",
  "aggregate": "sum",
  "attribute": "bid_request_cnt",
  "title": "Bid Request Count",
  "description":"Total Count of Bids"
},
```
Metrics can also be calculated by nesting various types of aggregates as shown in the example below. In this case, we are creating an Average Clear Price by dividing the sum of Clear Prices by the Sum of Impression Count. 

Operands include multiply, divide, add, subtract.

Note: we are also able to provide a prefix ($) to turn the amount into a currency.
```json
{
  "name": "avg_clear_price",
  "arithmetic": "divide",
  "operands": [
    {
     "aggregate": "sum",
     "attribute": "clear_price"
    },
    {
      "aggregate": "sum",
      "attribute": "imp_cnt"
    }
  ],
  "title": "Average Clear Price",
  "prefix": "$"
},
```
In addition to calculating metrics between fields, you can also use constants to adjust values. 

In this case, we're dividing Clear Price by 100 to calculate Gross Revenue. Another use case for constants is adjusting for any sampled data.
```json
{
  "name": "gross_revenue",
  "arithmetic": "divide",
  "operands": [
    {
       "aggregate": "sum",
       "attribute": "clear_price"
    },
    {
      "aggregate": "constant",
      "value": 100
    }
    ],
    "title": "Gross Revenue",
    "prefix": "$"
},
```

## Advanced Concepts: Metric Filtering

You can also filter results to only aggregate metrics if they meet certain conditions.

In the below example, Visited Websites is calculated as an aggregate as a sum of the CONVERTED field only when OUTCOME =  "Visited Website". 

The most common filters are: 
  - "is" (matches a single value) 
  - "in" (a list of values. **Note** update *value* to *values*)
```json
{
    "name": "VISITS",
    "filter": {
       "type": "is",
          "attribute": "OUTCOME",
          "value": [
              "Visited Website"
                ]
          },
    "aggregate": "sum",
    "attribute": "CONVERTED",
    "title": "Visits"
 },
```
In this example, we can exclude $0 bid_price to get a bid count sum.
```json
{
    "name": "count",
    "filter":{
        "type": "not",
        "attribute": "bid_price",
        "value": "0"
        },     
    "aggregate": "count",
    "attribute": "cnt",      
    "title": "Bid Count"
 },             

```
Multiple filters can also be applied using boolean logic (AND, OR). In the example below, "bids" are only aggregated when the "is_bid" field = 1 and "result" = "BID", "IMP".

**Note** Similar to the above, the top-level indicator for multiple filters is **filter** followed by the boolean (and, or) with the second plural **filters** applied.
```json
 {
    "name": "bids",
    "filter": {
      "type": "and",
      "filters": [
        {
          "type": "is",
          "attribute": "is_bid",
          "value": [
            "1"
          ]
        },
        {
          "type": "in",
          "attribute": "result",
          "values": [
            "BID",
            "IMP"
          ]
        }
      ]
    },
    "aggregate": "sum",
    "attribute": "cnt",
    "title": "Bids"
 },
```

:::danger Filtering when calculating averages
Note - filtering values when using average aggregates can have unexpected results. Instead, take a sum of the filtered metric divided by a count of the same filtered metric to achieve the desired average.
:::
