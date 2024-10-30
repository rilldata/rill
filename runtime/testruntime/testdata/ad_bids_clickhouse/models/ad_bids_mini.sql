select
    id,
    timestamp,
    publisher,
    domain,
    volume,
    impressions,
    clicks
from {{ ref "ad_bids_mini_source" }}
