SELECT
    (id::HUGEINT + 170141183460469231731687303715884105726)::HUGEINT as id,
    timestamp,
    publisher,
    domain,
    bid_price,
    volume,
    impressions,
    "ad words",
    clicks,
    1 as numeric_dim
FROM ad_bids_source
