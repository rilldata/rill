---
title: "Google Analytics Guide"
sidebar_label: "Google Analytics (GA4) Demo"
sidebar_position: 5
tags:
  - Tutorial
  - Quickstart
  - Example Project
---

# Google Analytics (GA4) Demo Guide

Rill makes it easy to visualize your GA4 data, empowering you to explore trends in website activity and make data-driven decisions with flexible, interactive analysis.

:::note prerequisites
Before you begin, ensure you have:
- [Exported your GA4 data to BigQuery](https://support.google.com/analytics/answer/9823238?hl=en#step2&zippy=%2Cin-this-article)
- [A service account key with access to BigQuery](https://cloud.google.com/bigquery/docs/use-service-accounts)

You can explore the dataset directly in BigQuery:
[GA4 Sample Dataset on BigQuery](https://console.cloud.google.com/bigquery?p=bigquery-public-data&d=ga4_obfuscated_sample_ecommerce&t=events_20210131&page=table&ref=ga4bigquery.com&inv=1&invt=Ab3HlA&project=rilldata)
:::

## Introduction

This guide will walk you through each step, from ingesting your GA4 data to building and exploring interactive dashboards in Rill.

---

## Step 1: Ingest Data

<img src='/img/guides/ga4/add-source.png' class='rounded-gif' />
<br />

Start by connecting to your GA4 data source. For this guide, we’ll use Google’s public GA4 dataset, which is available in BigQuery.

:::note Static Data
The data provided by Google is a static dataset with data from November 2020 to 2021.
:::

<img src='/img/guides/ga4/source-single-day.png' class='rounded-gif' />
<br />

When selecting BigQuery, Rill will prompt for:
```
`SQL` : select * from project_id.dataset.table; #select * from `bigquery-public-data.ga4_obfuscated_sample_ecommerce.events_20210131`
`project_id`: your google console project_id
`name`: name of source in Rill
```

### Partitioned GA4 Data

<img src='/img/guides/ga4/bigquery-partitioned.png' class='rounded-gif' />
<br />

If you are following along with this guide, use the following public dataset:

- **Project:** `bigquery-public-data`
- **Dataset:** `ga4_obfuscated_sample_ecommerce`
- **Table prefix:** `events_*`

If you are using your own data, you can make use of GA4's partitioned data using the `__TABLE_SUFFIX` parameter.

<img src='/img/guides/ga4/source-bigquery-suffix.png' class='rounded-gif' />
<br />

---

## Step 2: Expand the Columns

After connecting your data, you may want to expand nested columns or flatten the event structure for easier analysis by creating a model.

<img src='/img/guides/ga4/create-model.png' class='rounded-gif' />
<br />

:::tip Rill's Default Embedded Database, DuckDB
By default, Rill uses DuckDB as its underlying embedded database. This means that for any modeling, this will be done in DuckDB SQL.
:::

<img src='/img/guides/ga4/model.png' class='rounded-gif' />
<br />

<details>
<summary>Sample Model SQL</summary>

```SQL
SELECT
  event_date,
  user_pseudo_id,
  to_timestamp(event_timestamp/ 1000000) as event_timestamp, -- convert 1608921004450512 to a proper timestamp
  to_timestamp(user_first_touch_timestamp / 1000000) as user_first_touch_timestamp, --convert 1608921004450512 to timestamp, used to count time of session later

  stream_id,
  platform,
  event_name,
  event_value_in_usd, 
  event_bundle_sequence_id, 
  
  -- privacy_info
  -- json_extract(main.privacy_info, '$.ads_storage')::VARCHAR AS priv_info_ads_storage,
  -- json_extract(main.privacy_info, '$.analytics_storage')::VARCHAR AS priv_info_analytics_storage,
  -- json_extract(main.privacy_info, '$.uses_transient_token')::VARCHAR AS priv_info_uses_transient_token,
  
  -- user_ltv
  json_extract(main.user_ltv, '$.currency')::VARCHAR AS ltv_currency,
  CAST(json_extract(main.user_ltv, '$.revenue') AS DOUBLE) AS ltv_revenue,
  
  -- device
  -- json_extract(main.device, '$.advertising_id')::VARCHAR AS device_advertising_id,
  json_extract(main.device, '$.category')::VARCHAR AS device_category,
  -- json_extract(main.device, '$.is_limited_ad_tracking')::VARCHAR AS device_is_limited_ad_tracking,
  json_extract(main.device, '$.language')::VARCHAR AS device_language,
  json_extract(main.device, '$.mobile_brand_name')::VARCHAR AS device_mobile_brand_name,
  -- json_extract(main.device, '$.mobile_marketing_name')::VARCHAR AS device_mobile_marketing_name,
  json_extract(main.device, '$.mobile_model_name')::VARCHAR AS device_mobile_model_name,
  -- json_extract(main.device, '$.mobile_os_hardware_model')::VARCHAR AS device_mobile_os_hardware_model,
  json_extract(main.device, '$.operating_system')::VARCHAR AS device_operating_system,
  json_extract(main.device, '$.operating_system_version')::VARCHAR AS device_operating_system_version,
  -- CAST(json_extract(main.device, '$.time_zone_offset_seconds') AS INT) AS device_time_zone_offset_seconds,
  -- json_extract(main.device, '$.vendor_id')::VARCHAR AS device_vendor_id,
  json_extract(main.device, '$.web_info.browser')::VARCHAR AS device_browser,
  json_extract(main.device, '$.web_info.browser_version')::VARCHAR AS device_browser_version,
  
  -- geo
  json_extract(main.geo, '$.city')::VARCHAR AS geo_city,
  json_extract(main.geo, '$.continent')::VARCHAR AS geo_continent,
  json_extract(main.geo, '$.country')::VARCHAR AS geo_country,
  -- json_extract(main.geo, '$.metro')::VARCHAR AS geo_metro,
  json_extract(main.geo, '$.region')::VARCHAR AS geo_region,
  json_extract(main.geo, '$.sub_continent')::VARCHAR AS geo_sub_continent,
  
  -- traffic_source
  json_extract(main.traffic_source, '$.medium')::VARCHAR AS traffic_medium,
  json_extract(main.traffic_source, '$.name')::VARCHAR AS traffic_name,
  json_extract(main.traffic_source, '$.source')::VARCHAR AS traffic_source,
  
  -- ecommerce
  CAST(json_extract(main.ecommerce, '$.purchase_revenue') AS DOUBLE) AS ecommerce_purchase_revenue,
  CAST(json_extract(main.ecommerce, '$.purchase_revenue_in_usd') AS DOUBLE) AS ecommerce_purchase_revenue_in_usd,
  CAST(json_extract(main.ecommerce, '$.refund_value') AS DOUBLE) AS ecommerce_refund_value,
  CAST(json_extract(main.ecommerce, '$.refund_value_in_usd') AS DOUBLE) AS ecommerce_refund_value_in_usd,
  CAST(json_extract(main.ecommerce, '$.shipping_value') AS DOUBLE) AS ecommerce_shipping_value,
  CAST(json_extract(main.ecommerce, '$.shipping_value_in_usd') AS DOUBLE) AS ecommerce_shipping_value_in_usd,
  CAST(json_extract(main.ecommerce, '$.tax_value') AS DOUBLE) AS ecommerce_tax_value,
  CAST(json_extract(main.ecommerce, '$.tax_value_in_usd') AS DOUBLE) AS ecommerce_tax_value_in_usd,
  -- CAST(json_extract(main.ecommerce, '$.total_item_quantity') AS INT) AS ecommerce_total_item_quantity,
  json_extract(main.ecommerce, '$.transaction_id')::VARCHAR AS ecommerce_transaction_id,
  -- CAST(json_extract(main.ecommerce, '$.unique_items') AS INT) AS ecommerce_unique_items,
  
-- items (exploded)
-- json_extract(item.value, '$.affiliation')::VARCHAR AS value_item_affiliation,
-- json_extract(item.value, '$.coupon')::VARCHAR AS value_item_coupon,
json_extract(item.value, '$.creative_name')::VARCHAR AS value_creative_name,
-- json_extract(item.value, '$.creative_slot')::VARCHAR AS value_creative_slot,
json_extract(item.value, '$.item_brand')::VARCHAR AS value_item_brand,
json_extract(item.value, '$.item_category')::VARCHAR AS value_item_category,
-- json_extract(item.value, '$.item_category2')::VARCHAR AS value_item_category2,
-- json_extract(item.value, '$.item_category3')::VARCHAR AS value_item_category3,
-- json_extract(item.value, '$.item_category4')::VARCHAR AS value_item_category4,
-- json_extract(item.value, '$.item_category5')::VARCHAR AS value_item_category5,
json_extract(item.value, '$.item_id')::VARCHAR AS value_item_id,
-- json_extract(item.value, '$.item_list_id')::VARCHAR AS value_item_list_id,
json_extract(item.value, '$.item_list_index')::VARCHAR AS value_item_list_index,
json_extract(item.value, '$.item_list_name')::VARCHAR AS value_item_list_name,
json_extract(item.value, '$.item_name')::VARCHAR AS value_item_name,
CAST(json_extract(item.value, '$.item_refund') AS DOUBLE) AS value_item_refund,
CAST(json_extract(item.value, '$.item_refund_in_usd') AS DOUBLE) AS value_item_refund_in_usd,
CAST(json_extract(item.value, '$.item_revenue') AS DOUBLE) AS value_item_revenue,
CAST(json_extract(item.value, '$.item_revenue_in_usd') AS DOUBLE) AS value_item_revenue_in_usd,
json_extract(item.value, '$.item_variant')::VARCHAR AS value_item_variant,
-- json_extract(item.value, '$.location_id')::VARCHAR AS value_location_id,
CAST(json_extract(item.value, '$.price') AS DOUBLE) AS value_item_price,
CAST(json_extract(item.value, '$.price_in_usd') AS DOUBLE) AS value_item_price_in_usd,
-- json_extract(item.value, '$.promotion_id')::VARCHAR AS value_promotion_id,
json_extract(item.value, '$.promotion_name')::VARCHAR AS value_promotion_name,
-- CAST(json_extract(item.value, '$.quantity') AS INT) AS value_item_quantity

FROM obfuscated_sample_ecommerce AS main,
json_each(main.items) AS item
```
</details>

---

## Step 3: Create a Metrics View

With your data modeled, you can now define metrics and dimensions for analysis. In Rill, you have the option to create metrics via AI, which will send the schema and related information to OpenAI to build a basic metrics view. From there, you can make changes, add, or remove unneeded dimensions and measures.

<img src='/img/guides/ga4/create-metrics-with-ai.png' class='rounded-gif' />
<br />

Leverage Rill’s AI-powered features to quickly generate metrics views tailored to your business questions.

<img src='/img/guides/ga4/metrics-view.png' class='rounded-gif' />
<br />

---

## Step 4: Visualize Your Data

Finally, explore your GA4 data with interactive dashboards and visualizations. Here are a few example analyses you can perform in Rill:

- **Traffic Trends:** Track daily or weekly website sessions and identify spikes or drops in user activity.
- **User Segmentation:** Break down users by country, device type, or traffic source to understand your audience.
- **Ecommerce Performance:** Analyze purchase revenue, refund rates, and top-selling products over time.
- **Event Analysis:** Visualize key events (like sign-ups or checkouts) to optimize conversion funnels.

<img src='/img/guides/ga4/explore-dashboard.png' class='rounded-gif' />
<br />

Use Rill’s dashboard to analyze trends, segment users, and uncover insights from your GA4 events.

---