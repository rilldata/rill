select
    *,
    -- Add an offset timestamp that is 7 days earlier than the primary timestamp
    "timestamp" - INTERVAL '7 days' as offset_timestamp
from AdBids
