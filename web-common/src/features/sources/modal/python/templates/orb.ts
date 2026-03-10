export const script = `import os
import requests
import pandas as pd

output_path = os.environ["RILL_OUTPUT_PATH"]
api_key = os.environ["ORB_API_KEY"]

headers = {"Authorization": f"Bearer {api_key}"}
base_url = "https://api.withorb.com/v1"

# Fetch subscriptions (paginated)
subscriptions = []
cursor = None
while True:
    params = {"limit": 50}
    if cursor:
        params["cursor"] = cursor
    resp = requests.get(f"{base_url}/subscriptions", headers=headers, params=params)
    resp.raise_for_status()
    data = resp.json()
    for sub in data.get("data", []):
        subscriptions.append({
            "id": sub["id"],
            "customer_id": sub.get("customer", {}).get("id"),
            "plan_id": sub.get("plan", {}).get("id"),
            "status": sub.get("status"),
            "start_date": sub.get("start_date"),
            "end_date": sub.get("end_date"),
        })
    cursor = data.get("pagination_metadata", {}).get("next_cursor")
    if not cursor:
        break

df = pd.DataFrame(subscriptions)
df.to_parquet(output_path, index=False)
print(f"Wrote {len(df)} subscriptions to {output_path}")
`;
