---
title: Metrics View Example YAML
tags:
- metrics
- code
- complete_file
- duckdb
docs: https://docs.rilldata.com/build/metrics-view
hash: 08c9d6cabf756ed9d67b078054fa40bf4e41e1a4daf41f3a1e5d783ae43a6967
---

```YAML
# Metrics view YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/dashboards

version: 1
model: "bids_data_model"
type: metrics_view

timeseries: "__time"
smallest_time_grain: "hour"

measures:
  - display_name: Advertising Spend Overall
    name: overall_spend
    expression: sum(media_spend_usd) / 1000
    description: Total Spend
    format_preset: currency_usd
  - display_name: "Bids"
    name: total_bids
    expression: "sum(bid_cnt)"
    description: "Total Bids"
    format_preset: humanize
  - display_name: "Impressions"
    name: impressions
    expression: "sum(imp_cnt)"
    description: "Total Impressions"
    format_preset: humanize
  - display_name: "Win Rate"
    name: win_rate
    expression: "sum(imp_cnt)*1.0 / nullif(sum(bid_cnt), 0)"
    description: "Impressions / Bids"
    format_preset: percentage
  - display_name: "Clicks"
    name: clicks
    expression: "sum(click_reg_cnt)"
    description: "Total Clicks"
    format_preset: humanize
  - display_name: "CTR"
    name: ctr
    expression: "sum(click_reg_cnt)*1.0 / nullif(sum(imp_cnt), 0)"
    description: "Click Through Rate"
    format_preset: percentage
  - display_name: "Video Starts"
    name: video_starts
    expression: "sum(video_start_cnt)"
    description: "Total Video Starts"
    format_preset: humanize
  - display_name: "Video Completes"
    name: video_completes
    expression: "sum(video_complete_cnt)"
    description: "Total Video Completes"
    format_preset: humanize
  - display_name: "Video Completion Rate"
    name: video_completion_rate
    expression: "sum(video_complete_cnt)*1.0 / nullif(sum(video_start_cnt), 0)"
    description: "Video Completion Rate"
    format_preset: percentage
  - display_name: "Avg Bid Price"
    name: avg_bid_price
    expression: "sum(bid_price_usd)*1.0/ nullif(sum(bid_cnt)/1000, 0)"
    description: "Average Bid Price"
    format_preset: currency_usd
  - display_name: "eCPM"
    name: ecpm
    expression: "sum(media_spend_usd)*1.0 / 1000 / nullif(sum(imp_cnt), 0)"
    description: "eCPM"
    format_preset: currency_usd
  - display_name: "Avg Bid Floor"
    name: avg_bid_floor
    expression: "sum(bid_floor)*1.0 / nullif(sum(has_bid_floor_cnt), 0)"
    description: "Average Bid Floor"
    format_preset: currency_usd
  - name: bids_7day_rolling_avg
    display_name: 7 Day Bid rolling avg
    expression: AVG(total_bids)
    requires: [total_bids]
    window:
      order: "__time"
      frame: RANGE BETWEEN INTERVAL 6 DAY PRECEDING AND CURRENT ROW


dimensions:
  - column: adomain
    name: adomain
    display_name: Adomain
  - column: advertiser_name
    name: advertiser_name
    display_name: Advertiser Name
  - display_name: App or Site
    name: app_or_site
    column: app_or_site
    description: ""
  - display_name: Site Domain
    name: sites_domain
    column: app_site_domain
    description: ""
  - column: app_site_name
    name: app_site_name
    display_name: App Name
  - name: auction_type
    display_name: Auction Type
    column: auction_type
    description: ""
  - display_name: Bid Floor Bucket
    column: bid_floor_bucket
    description: ""
    name: bid_floor_bucket
  - name: campaign_name
    display_name: Campaign Name
    column: campaign_name
    description: ""
  - name: creative_type
    display_name: Creative type
    column: creative_type
    description: ""
  - name: device_os
    display_name: Device OS
    column: device_os
    description:
  - name: device_osv
    display_name: Device OSV
    column: device_osv
    description: ""
  - name: device_region
    display_name: Device Region
    column: device_region
    description: ""
  - display_name: Device Type
    column: device_type
    description: ""
    name: device_type
  - name: interstitial
    display_name: Interstitial
    column: interstitial
    description: ""
  - name: line_item_name
    display_name: Line Item Name
    column: line_item_name
    description: ""
  - display_name: Placement Type
    column: placement_type
    description: ""
    name: placement_type
  - name: platform_browser
    display_name: Platform Browser
    column: platform_browser
    description: ""
  - name: player_size
    display_name: Player Size
    column: player_size
    description: ""
  - name: privacy
    display_name: Privacy
    column: privacy
    description: ""
  - name: pub_name
    display_name: Pub Name
    column: pub_name
    description: ""
  - name: sdk
    display_name: SDK
    column: sdk
    description: ""
  - name: video_activity
    display_name: Video Activity
    column: video_activity
    description: ""
```
