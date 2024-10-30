select
    id,
    timestamp,
    publisher,
    domain,
    bid_price AS bid_prices
from {{ ref "ad_bids_source" }}
