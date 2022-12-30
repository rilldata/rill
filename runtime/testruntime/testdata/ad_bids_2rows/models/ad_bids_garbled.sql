SELECT
    (id::HUGEINT + 170141183460469231731687303715884105726)::HUGEINT as id,
    timestamp as "timestamp name with spaces",
    publisher,
    domain,
    bid_price,
    volume,
    impressions,
    "ad words",
    clicks
FROM ad_bids_source
