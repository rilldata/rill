package testruntime

import "fmt"

// ProjectOpenRTB returns project files that resemble our public rill-openrtb-prog-ads example project.
// Unlike the example project, it:
// 1. uses local cached copies of the datasets for faster and more reliable tests.
// 2. has been trimmed to the essential resources and properties.
func ProjectOpenRTB(t TestingT) (int, map[string]string) {
	auctionsParquet := DatasetPath(t, DatasetOpenRTBAuctions)
	bidsParquet := DatasetPath(t, DatasetOpenRTBBids)

	return 9, map[string]string{
		// Raw auctions data (NOTE: not materialized)
		"models/auctions_data_raw.yaml": fmt.Sprintf(`
type: model
materialize: false
sql: SELECT * FROM read_parquet('%s')
`, auctionsParquet),
		// Raw bids data (NOTE: not materialized)
		"models/bids_data_raw.yaml": fmt.Sprintf(`
type: model
materialize: false
sql: SELECT * FROM read_parquet('%s')
`, bidsParquet),
		// Cleaned auctions data (NOTE: materialized)
		"models/auctions_data.yaml": `
type: model
materialize: true
sql: |
  SELECT
    * EXCLUDE (device_region),
    CASE WHEN device_region ILIKE '%/%' THEN SPLIT(device_region, '/')[2] ELSE 'Unknown' END AS device_state,
    CASE WHEN device_region ILIKE '%/%' THEN SPLIT(device_region, '/')[1] ELSE 'Unknown' END AS device_country
  FROM auctions_data_raw
  `,
		// Cleaned bids data (NOTE: materialized)
		"models/bids_data.yaml": `
type: model
materialize: true
sql: |
  SELECT * FROM bids_data_raw
`,
		// Auctions metrics
		"metrics/auctions_metrics.yaml": `
type: metrics_view

model: auctions_data
timeseries: __time
smallest_time_grain: hour

dimensions:
  - column: app_site_name
  - column: app_site_domain
  - column: pub_name
  - column: app_site_cat
  - column: ad_size
  - column: device_state
  - column: device_osv
  - column: bid_floor_bucket
  - column: platform_browser
  - column: device_os
  - column: device_type
  - column: ad_position
  - column: video_max_duration_bucket
  - column: video_min_duration_bucket
  - column: placement_type
  - column: auction_type
  - column: app_or_site
  - column: device_country

measures:
  - name: requests
    expression: sum(bid_request_cnt)
    format_preset: humanize
  - name: avg_bid_floor
    expression: sum(bid_floor) / sum(has_bid_floor_cnt)
    format_preset: currency_usd
  - name: 1d_qps
    expression: sum(bid_request_cnt) / 86400
    format_preset: humanize
`,
		// Bids metrics
		"metrics/bids_metrics.yaml": `
type: metrics_view

model: bids_data
timeseries: __time
smallest_time_grain: hour

dimensions:
  - column: adomain
  - column: advertiser_name
  - column: app_or_site
  - column: app_site_domain
  - column: app_site_name
  - column: auction_type
  - column: bid_floor_bucket
  - column: campaign_name
  - column: creative_type
  - column: device_os
  - column: device_osv
  - column: device_region
  - column: device_type
  - column: interstitial
  - column: line_item_name
  - column: placement_type
  - column: platform_browser
  - column: player_size
  - column: privacy
  - column: pub_name
  - column: sdk
  - column: video_activity

measures:
  - name: overall_spend
    expression: sum(media_spend_usd)/1000
    description: Total Spend
    format_preset: currency_usd
  - name: total_bids
    expression: sum(bid_cnt)
    description: Total Bids
    format_preset: humanize
  - name: impressions
    expression: sum(imp_cnt)
    description: Total Impressions
    format_preset: humanize
  - name: win_rate
    expression: sum(imp_cnt)*1.0/sum(bid_cnt)
    description: Impressions / Bids
    format_preset: percentage
  - name: clicks
    expression: sum(click_reg_cnt)
    description: Total Clicks
    format_preset: humanize
  - name: ctr
    expression: sum(click_reg_cnt)*1.0/nullif(sum(imp_cnt),0)
    description: Click Through Rate
    format_preset: percentage
  - name: video_starts
    expression: sum(video_start_cnt)
    description: Total Video Starts
    format_preset: humanize
  - name: video_completes
    expression: sum(video_complete_cnt)
    description: Total Video Completes
    format_preset: humanize
  - name: video_completion_rate
    expression: sum(video_complete_cnt)*1.0/sum(video_start_cnt)
    description: Video Completion Rate
    format_preset: percentage
  - name: avg_bid_price
    expression: sum(bid_price_usd)*1.0/sum(bid_cnt)/1000
    description: Average Bid Price
    format_preset: currency_usd
  - name: ecpm
    expression: sum(media_spend_usd)*1.0/1000/nullif(sum(imp_cnt),0)
    description: eCPM
    format_preset: currency_usd
  - name: avg_bid_floor
    expression: sum(bid_floor)*1.0/sum(has_bid_floor_cnt)
    description: Average Bid Floor
    format_preset: currency_usd    
  - name: bids_7day_rolling_avg
    expression: AVG(total_bids)
    requires: [total_bids]
    window:
      order: __time
      frame: RANGE BETWEEN INTERVAL 6 DAY PRECEDING AND CURRENT ROW
`,
	}
}
