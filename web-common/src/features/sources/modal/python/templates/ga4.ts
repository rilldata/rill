export const script = `import os
import pandas as pd
from google.analytics.data_v1beta import BetaAnalyticsDataClient
from google.analytics.data_v1beta.types import (
    DateRange, Dimension, Metric, RunReportRequest,
)

property_id = "YOUR_PROPERTY_ID"  # Replace with your GA4 property ID
output_path = os.environ["RILL_OUTPUT_PATH"]

client = BetaAnalyticsDataClient()

request = RunReportRequest(
    property=f"properties/{property_id}",
    dimensions=[
        Dimension(name="date"),
        Dimension(name="country"),
        Dimension(name="sessionDefaultChannelGroup"),
    ],
    metrics=[
        Metric(name="sessions"),
        Metric(name="totalUsers"),
        Metric(name="screenPageViews"),
        Metric(name="bounceRate"),
    ],
    date_ranges=[DateRange(start_date="90daysAgo", end_date="today")],
)

response = client.run_report(request)

rows = []
for row in response.rows:
    rows.append({
        "date": row.dimension_values[0].value,
        "country": row.dimension_values[1].value,
        "channel_group": row.dimension_values[2].value,
        "sessions": int(row.metric_values[0].value),
        "total_users": int(row.metric_values[1].value),
        "page_views": int(row.metric_values[2].value),
        "bounce_rate": float(row.metric_values[3].value),
    })

df = pd.DataFrame(rows)
df["date"] = pd.to_datetime(df["date"], format="%Y%m%d")
df.to_parquet(output_path, index=False)
print(f"Wrote {len(df)} rows to {output_path}")
`;
