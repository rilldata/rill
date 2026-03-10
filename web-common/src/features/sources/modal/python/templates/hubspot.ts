export const script = `import os
import requests
import pandas as pd

output_path = os.environ["RILL_OUTPUT_PATH"]
access_token = os.environ["HUBSPOT_ACCESS_TOKEN"]

headers = {
    "Authorization": f"Bearer {access_token}",
    "Content-Type": "application/json",
}

contacts = []
after = None

while True:
    url = "https://api.hubapi.com/crm/v3/objects/contacts"
    params = {"limit": 100, "properties": "firstname,lastname,email,company,lifecyclestage,createdate"}
    if after:
        params["after"] = after
    resp = requests.get(url, headers=headers, params=params)
    resp.raise_for_status()
    data = resp.json()
    for record in data["results"]:
        props = record.get("properties", {})
        contacts.append({
            "id": record["id"],
            "first_name": props.get("firstname"),
            "last_name": props.get("lastname"),
            "email": props.get("email"),
            "company": props.get("company"),
            "lifecycle_stage": props.get("lifecyclestage"),
            "created_at": props.get("createdate"),
        })
    paging = data.get("paging")
    if paging and "next" in paging:
        after = paging["next"]["after"]
    else:
        break

df = pd.DataFrame(contacts)
if "created_at" in df.columns:
    df["created_at"] = pd.to_datetime(df["created_at"], format="ISO8601")
df.to_parquet(output_path, index=False)
print(f"Wrote {len(df)} contacts to {output_path}")
`;
