---
title: "OpenRTB Analytics Demo"
sidebar_label: "OpenRTB Analytics Demo"
hide_table_of_contents: false
sidebar_position: 20
tags:
    - Tutorial
    - Quickstart
---

# OpenRTB Analytics Demo

Learn how to analyze real-time bidding (RTB) data with Rill using the OpenRTB Analytics demo project. This guide shows you how to track ad impressions, bids, wins, and revenue across different publishers, advertisers, and campaigns.

## Step 1: Clone the Project

```bash
# Clone the OpenRTB Analytics demo
git clone https://github.com/rilldata/rill-examples.git
cd rill-examples/rill-openrtb-prog-ads

# Start Rill Developer
rill start
```

Visit [http://localhost:9009](http://localhost:9009) to explore your OpenRTB analytics dashboard.

## Step 2: Project Structure

The project is organized as follows:

```
rill-openrtb-prog-ads/
├── rill.yaml                           # Project configuration
├── sources/                            # Data source definitions
│   ├── auction_data_raw.yaml           # Ad impression data
│   └── bids_data_raw.yaml.yaml         # Bid request/response data
├── models/                             # SQL transformations
│   ├── auction_data_model.sql          # Workaround SQL to keep data up to date
│   └── bids_data_model.sql             # Workaround SQL to keep data up to date
├── metrics/                            # Defined measures and dimensions
│   ├── auction_metrics.yaml            # Auction metrics
│   └── bid_metrics.yaml                # Bids metrics
├── dashboards/                         # Dashboard configurations
│   ├── auction_explore.yaml            # Auction Explore dashboard
│   ├── bids_explore.yaml               # Bid Explore dashboard
│   └── executive_overview.yaml         # Canvas dashboard
└── README.md                           # Project documentation
```

## Step 3: Data Sources

### Bids Source

The bids source captures bid request and response data:

```yaml
# sources/bids_data_raw.yaml
type: source
connector: "https"
uri: "https://storage.googleapis.com/rilldata-public/auction_data.parquet"
```

**What this does:**
- Records all bid requests and responses
- Tracks bid prices and win/loss status
- Enables analysis of bidding patterns and market dynamics

### Auction Source

The auction source tracks aggregated auction metrics and bid outcomes:

```yaml
# sources/auction_data_raw.yaml
type: source
connector: "https"
uri: "https://storage.googleapis.com/rilldata-public/auction_data.parquet"
```

**What this does:**
- Connects to the `rilldata-public` GCS bucket
- Fetches aggregated auction data with key metrics
- Tracks bid requests, responses, wins, and bid floors
- Data includes volume metrics, performance ratios, and revenue data


## Step 4: Data Models

In this case, we are not joining the data models and having two unique explore models and joining the visualization in a [Canvas Dashboard](/build/dashboards/canvas).


## Step 5: Creating your Metrics View

Metrics in Rill define the measures and dimensions that power your RTB dashboards:

```yaml
# Metrics view YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/metrics-views

version: 1
model: "auction_data_model"
type: metrics_view

timeseries: "__time"
smallest_time_grain: "hour"

measures:
  - display_name: Total Requests
    name: requests
    expression: sum(bid_request_cnt)
    description: Total Requests
    format_preset: humanize
  - display_name: "Avg Bid Floor"
    name: avg_bid_floor
    expression: "sum(bid_floor) / sum(has_bid_floor_cnt) "
    description: "Average Bid Floor"
    format_preset: currency_usd
  - display_name: "1D QPS"
    name: 1d_qps
    expression: "sum(bid_request_cnt) / 86400 "
    description: "1D QPS"
    format_preset: humanize


dimensions:
  - name: app_site_name
    display_name: App Site Name
    column: app_site_name
  - display_name: App Site Domain
    column: app_site_domain
    name: app_site_domain
  - name: pub_name
    display_name: Pub Name
    column: pub_name
  - name: app_site_cat
    display_name: App Site Cat
    column: app_site_cat
  - name: ad_size
    display_name: Ad Size
    column: ad_size
  - name: device_state
    display_name: Device State
    column: device_state
  - name: device_osv
    display_name: Device OS Version
    column: device_osv
  - name: bid_floor_bucket
    display_name: Bid Floor Bucket
    column: bid_floor_bucket
  - name: platform_browser
    display_name: Platform Browser
    column: platform_browser
  - name: device_os
    display_name: Device OS
    column: device_os
  - name: device_type
    display_name: Device Type
    column: device_type
  - name: ad_position
    display_name: Ad Position
    column: ad_position
  - name: video_max_duration_bucket
    display_name: Video Max Duration Bucket
    column: video_max_duration_bucket
  - name: video_min_duration_bucket
    display_name: Video Min Duration Bucket
    column: video_min_duration_bucket
  - display_name: Placement Type
    column: placement_type
    name: placement_type
  - name: auction_type
    display_name: Auction Type
    column: auction_type
  - name: app_or_site
    display_name: App or Site
    column: app_or_site
  - name: device_country
    display_name: "Device Country"
    column: "device_country"

ai_instructions: ...
```

**What this metric view does:**

- **Measures** define the key RTB performance indicators:
  - `Total Requests` - Total number of bid requests processed
  - `Avg Bid Floor` - Average minimum acceptable price across auctions
  - `1D QPS` - Daily queries per second (bid request volume)
  - `avg_profit_margin_pct` - Average profit margin percentage

- **Dimensions** enable analysis across different segments:
  - `app_site_name` - Application or website name for publisher analysis
  - `pub_name` - Publisher name for performance tracking
  - `ad_size` - Ad unit dimensions (300x250, 728x90, etc.)
  - `bid_floor_bucket` - Grouped bid floor ranges for optimization analysis
  - `device_country` - Geographic performance by country
  - `device_type` - Mobile, desktop, or tablet device analysis
  - `auction_type` - Type of auction (first-price, second-price, etc.)
  - `placement_type` - Ad placement context (banner, video, native)




## Step 6: Dashboard Exploration

#### **Features**
- **Explore Slice-and-Dice** - For data exploration and ad-hoc analysis
- **Canvas** - Traditional charts and visualizations
- **Pivot/Flat Table** - Tabular data views with sorting and grouping
- **Measure's TDD** - Granular Analysis of single measure

#### **Selectors**
- **Date range selector** - Analyze specific time periods
- **Time Comparison Toggle** - Compare previous time periods
- **Dimensions Comparison** - Compare unique dimension values

#### **Filters**
- **Publisher filter** - Focus on specific publishers
- **Advertiser filter** - Analyze specific advertisers
- **Geographic filter** - Regional performance analysis



### Let's answer some basic questions!


For campaign `Leafly_MarketMight`, which site domains are getting the most bids? 
<img src='/img/tutorials/quickstart/openrtb-analytics-1.png' class='rounded-gif'/>
<br />

Which US region has the highest activity and what devices are being used?
<img src='/img/tutorials/quickstart/openrtb-analytics-2.png' class='rounded-gif'/>
<br />

Compare Bids and Auctions.
<img src='/img/tutorials/quickstart/openrtb-analytics-3.png' class='rounded-gif'/>
<br />


These are just some of the insights that you can find within your explore dashboard but you'll find more hidden gems in your data as you continue to use Rill. Please let us know if you have any other questions!