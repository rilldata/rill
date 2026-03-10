export const script = `import os
import pandas as pd

output_path = os.environ["RILL_OUTPUT_PATH"]

# Your data extraction logic here
data = [{"column": "value"}]

df = pd.DataFrame(data)
df.to_parquet(output_path, index=False)
print(f"Wrote {len(df)} rows to {output_path}")
`;
