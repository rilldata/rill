-- Model SQL
-- Reference documentation: https://docs.rilldata.com/build/models
-- @materialize: true
select
    * EXCLUDE (device_region),
    CASE
        WHEN device_region ILIKE '%/%' THEN SPLIT(device_region, '/') [2]
        ELSE 'Unknown'
    END AS device_state,
    CASE
        WHEN device_region ILIKE '%/%' THEN SPLIT(device_region, '/') [1]
        ELSE 'Unknown'
    END AS device_country
from
    auction_data_raw