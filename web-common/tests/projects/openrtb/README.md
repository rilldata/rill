# OpenRTB Programmatic Advertising

This is a demo project designed to illustrate using Rill to analyze programmatic bid logs using the canonical open RTB framework. 

If you have added the full Rill Example project, run `rill start` from this directory to get started.

To run this example specifically:

```
git clone https://github.com/rilldata/rill-examples.git
cd rill-examples/rill-openrtb-prog-ads
rill start
```

Rill will build your project from data sources to dashboard and then launch in a new browser window.

## Overview
This dataset contains a week of sampled programmatic bid stream data in two data sources - Auctions and Bids. 

Advertisers, DSPs, SSPs, and Publishers will all recognize the familiar metrics (auctions, bids, wins, bid price, bid floor) and dimensions (domain, device details, app/site, etc). Rill’s was born out of a long history with programmatic data via Metamarkets and is well-suited for this type of analysis. More details on OpenRTB via the IAB: https://iabtechlab.com/standards/openrtb/.

## Data Model
In these datasets, you’ll see:

Auction Data:
  - Illustrative Bid Requests sent to advertisers for programmatic bidding 

Bid Data: 
  - Illustrative Bid Responses to those requests including bid prices, winning bids, and advertiser information

## Dashboard Details

For Buyers:
  - Manage all campaigns across multiple supply sources
  - View inventory and audience availability to avoid missing key opportunities and to optimize spend

For Sellers:
  - See both direct and indirect channels across your digital assets
  - Quickly slice and dice inventory to find trends and discover revenue opportunities

For Marketplaces/Technology Providers:
  - Troubleshoot campaigns and quickly identify ad server issues
  - Instantly view top-line revenue, volume, eCPM, and other key metrics without pulling complex reports

## Extra Dashboard

An additional dashboard is created with row policies enabled for specific emails. This is used in our embed examples found, [here](https://rill-embedding-example.netlify.app/).