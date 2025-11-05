---
title: Explore With Defaults Example YAML
tags:
- explore
- code
- complete_file
docs: https://docs.rilldata.com/explore
hash: 215124a6f66200f8b99e4a31d64740d4308a450218802f16e80f6d6fa2f2a5c5
---

```YAML
# Explore YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/explores

type: explore

display_name: "Programmatic Ads Bids"
metrics_view: bids_metrics

theme: BrandedTheme

dimensions: '*'
measures: '*'

time_zones:
  - America/New_York
  - Europe/London
  - Europe/Paris
  - Asia/Jerusalem
  - Europe/Moscow
  - Asia/Tokyo
  - Australia/Sydney

defaults:
  time_range: P7D
  measures:
    - overall_spend
    - total_bids
    - impressions
    - win_rate
    - clicks
    - ctr
    - video_completion_rate
    - avg_bid_price
    - ecpm
    - avg_bid_floor
  dimensions:
    - adomain
    - advertiser_name
    - app_or_site
    - sites_domain
    - app_site_name
    - auction_type
    - bid_floor_bucket
    - campaign_name
    - creative_type
    - device_os
    - device_osv
```
