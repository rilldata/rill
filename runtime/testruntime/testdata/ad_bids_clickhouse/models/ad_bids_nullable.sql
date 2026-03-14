SELECT
    toNullable(publisher) AS publisher,
    toNullable(domain) AS domain,
    timestamp
FROM {{ ref "ad_bids_source" }}
