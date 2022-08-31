---
title: "Edit Dashboard Metrics"
slug: "explore-json-examples"
excerpt: "Illustrative snippets for updating your dashboards"
hidden: false
createdAt: "2021-06-16T23:59:27.521Z"
updatedAt: "2022-07-13T07:13:45.174Z"
---
[block:api-header]
{
  "title": "Adding a Dashboard in Explore"
}
[/block]
To get started, a Staging dashboard will be created for your each datasource in your account. These staging dashboards will contain a full set of dimensions and metrics available in your datasource.

We recommend starting with a single dashboard per datasource that contains all metrics and dimensions. If there are common views of the data, [bookmarks are a helpful tool](https://enterprise.rilldata.com/docs/bookmarking) to save a specific set of filters, metrics and dimensions.

To add or edit dashboards, please contact your TAM or email [support@rilldata.com](mailto:support@rilldata.com) to make sure Admin capabilities are available for your account.
[block:api-header]
{
  "title": "Working with Dimensions"
}
[/block]
Each Dimension definition contains a short list of fields:

  * a "bucket" (the type of dimension - 99% of cases will be "identity")
  * declared name (the name in the dashboard Datasource)
  * an attribute (the name in the Druid database)
  * a title (for both single and plural)
  * a description (optional)

 Each dimension (except for the last) is followed by a comma. 
[block:code]
{
  "codes": [
    {
      "code": " {\n  \"bucket\": \"identity\",\n  \"name\": \"api_frameworks\",\n  \"attribute\": \"api_frameworks\",\n  \"titleSingle\": \"API Framework\",\n  \"titlePlural\": \"API Frameworks\",\n  \"description\": \"Framework used by API\"\n},",
      "language": "json"
    }
  ]
}
[/block]
For dimensions with lookups, the JSON has a separate extraction function.

To add a lookup, a second statement is added including the type (registeredLookup), the lookup, and the ability to retain or remove missing values.

If the lookup is 1:1, set injective to true to improve performance.
[block:code]
{
  "codes": [
    {
      "code": "{\n  \"bucket\": \"identity\",\n  \"name\": \"device_type\",\n  \"attribute\": \"device_type\",\n  \"titleSingle\": \"Device Type\",\n  \"titlePlural\": \"Device Type\",\n  \"extractionFn\": {\n    \"type\": \"registeredLookup\",\n    \"lookup\": \"example_lookup_table\",\n    \"retainMissingValue\": true,\n    \"injective\": false,\n    \"optimize\": true\n    }\n},",
      "language": "json"
    }
  ]
}
[/block]
Beyond lookups to tables, lookup functions can also be used to replace values in a dimension with simple transform logic. 

Below the first, simpler example replaces missing values with "Not Available."

The second example replaces those Not Available values and also maps them to "False."
[block:code]
{
  "codes": [
    {
      "code": "{\n  \"bucket\": \"identity\",\n  \"name\": \"ad_network_qtl\",\n  \"attribute\": \"ad_network_id\",\n  \"titleSingle\": \"Ad Network Name\",\n  \"titlePlural\": \"Ad Network Name\",\n  \"description\": \"Ad Network that owns the package\",\n  \"extractionFn\": {\n    \"retainMissingValue\": false,\n    \"lookup\": \"ad_network_lookup\",\n    \"replaceMissingValueWith\": \"Not Available\",\n    \"type\": \"registeredLookup\",\n    \"optimize\": true,\n    \"injective\": false\n    }\n},\n {\n   \"bucket\": \"identity\",\n   \"name\": \"placement_flat_cpm_enabled\",\n   \"attribute\": \"placement_flat_cpm_enabled\",\n   \"titleSingle\": \"Flat CPM Placement\",\n   \"titlePlural\": \"Flat CPM Placement\",\n   \"extractionFn\": {\n      \"type\": \"lookup\",\n      \"dimension\": \"placement_flat_cpm_enabled\",\n      \"outputName\": \"placement_flat_cpm_enabled_val\",\n      \"replaceMissingValueWith\": \"False\",\n      \"retainMissingValue\": false,\n      \"lookup\": {\n        \"type\": \"map\",\n        \"map\": {\n          \"True\": \"True\",\n           \"False\": \"False\",\n           \"Not Available\": \"False\"\n           }\n       }\n    }\n}  ",
      "language": "json"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "Working with Metrics"
}
[/block]
Metrics passed directly within Druid are also straightforward and similar to basic dimensions by providing:

  *name
  *type of aggregate (sum, count, max, min, average)
  *attribute (database field to aggregate) 
  *title (display name) 
  *description (*optional mouseover*) 
[block:code]
{
  "codes": [
    {
      "code": "{\n  \"name\": \"bid_request_cnt\",\n  \"aggregate\": \"sum\",\n  \"attribute\": \"bid_request_cnt\",\n  \"title\": \"Bid Request Count\",\n  \"description\":\"Total Count of Bids\"\n},",
      "language": "json"
    }
  ]
}
[/block]
Metrics can also be calculated by nesting various types of aggregates as shown in the example below. In this case, we are creating an Average Clear Price by dividing the sum of Clear Prices by the Sum of Impression Count. 

Operands include multiply, divide, add, subtract.

Note: we are also able to provide a prefix ($) to turn the amount into a currency.
[block:code]
{
  "codes": [
    {
      "code": "{\n  \"name\": \"avg_clear_price\",\n  \"arithmetic\": \"divide\",\n  \"operands\": [\n    {\n     \"aggregate\": \"sum\",\n     \"attribute\": \"clear_price\"\n    },\n    {\n      \"aggregate\": \"sum\",\n      \"attribute\": \"imp_cnt\"\n    }\n  ],\n  \"title\": \"Average Clear Price\",\n  \"prefix\": \"$\"\n},",
      "language": "json"
    }
  ]
}
[/block]
In addition to calculating metrics between fields, you can also use constants to adjust values. 

In this case, we're dividing Clear Price by 100 to calculate Gross Revenue. Another use case for constants is adjusting for any sampled data.
[block:code]
{
  "codes": [
    {
      "code": "{\n  \"name\": \"gross_revenue\",\n  \"arithmetic\": \"divide\",\n  \"operands\": [\n    {\n       \"aggregate\": \"sum\",\n       \"attribute\": \"clear_price\"\n\t\t},\n    {\n      \"aggregate\": \"constant\",\n      \"value\": 100\n    }\n\t\t],\n\t\t\"title\": \"Gross Revenue\",\n    \"prefix\": \"$\"\n},",
      "language": "json"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "Advanced Concepts: Metric Filtering"
}
[/block]
You can also filter results to only aggregate metrics if they meet certain conditions.

In the below example, Visited Websites is calculated as an aggregate as a sum of the CONVERTED field only when OUTCOME =  "Visited Website". 

The most common filters are: 
  * "is" (matches a single value) 
  * "in" (a list of values. **Note** update *value* to *values*)
[block:code]
{
  "codes": [
    {
      "code": "{\n    \"name\": \"VISITS\",\n\t\t\"filter\": {\n       \"type\": \"is\",\n          \"attribute\": \"OUTCOME\",\n          \"value\": [\n              \"Visited Website\"\n                ]\n          },\n    \"aggregate\": \"sum\",\n    \"attribute\": \"CONVERTED\",\n    \"title\": \"Visits\"\n },",
      "language": "json"
    }
  ]
}
[/block]
In this example, we can exclude $0 bid_price to get a bid count sum.
[block:code]
{
  "codes": [
    {
      "code": "{\n    \"name\": \"count\",\n    \"filter\":{\n        \"type\": \"not\",\n        \"attribute\": \"bid_price\",\n        \"value\": \"0\"\n        },     \n    \"aggregate\": \"count\",\n    \"attribute\": \"cnt\",      \n    \"title\": \"Bid Count\"\n },             \n",
      "language": "json"
    }
  ]
}
[/block]
Multiple filters can also be applied using boolean logic (AND, OR). In the example below, "bids" are only aggregated when the "is_bid" field = 1 and "result" = "BID", "IMP".

**Note** Similar to the above, the top-level indicator for multiple filters is **filter** followed by the boolean (and, or) with the second plural **filters** applied.
[block:code]
{
  "codes": [
    {
      "code": " {\n          \"name\": \"bids\",\n          \"filter\": {\n            \"type\": \"and\",\n            \"filters\": [\n              {\n                \"type\": \"is\",\n                \"attribute\": \"is_bid\",\n                \"value\": [\n                  \"1\"\n                ]\n              },\n              {\n                \"type\": \"in\",\n                \"attribute\": \"result\",\n                \"values\": [\n                  \"BID\",\n                  \"IMP\"\n                ]\n              }\n            ]\n          },\n          \"aggregate\": \"sum\",\n          \"attribute\": \"cnt\",\n          \"title\": \"Bids\"\n },",
      "language": "json"
    }
  ]
}
[/block]

[block:callout]
{
  "type": "danger",
  "title": "Filtering when calculating averages",
  "body": "Note - filtering values when using average aggregates can have unexpected results. Instead, take a sum of the filtered metric divided by a count of the same filtered metric to achieve the desired average."
}
[/block]