export const script = `import os
import requests
import pandas as pd

output_path = os.environ["RILL_OUTPUT_PATH"]

# Replace with your API endpoint and auth
url = "https://api.example.com/data"
headers = {"Authorization": f"Bearer {os.environ.get('API_TOKEN', '')}"}

response = requests.get(url, headers=headers)
response.raise_for_status()
data = response.json()

df = pd.DataFrame(data)
df.to_parquet(output_path, index=False)
print(f"Wrote {len(df)} rows to {output_path}")
`;
