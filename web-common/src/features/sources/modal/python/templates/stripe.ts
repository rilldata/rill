export const script = `import os
import stripe
import pandas as pd

output_path = os.environ["RILL_OUTPUT_PATH"]
stripe.api_key = os.environ["STRIPE_API_KEY"]

charges = []
has_more = True
starting_after = None

while has_more:
    params = {"limit": 100}
    if starting_after:
        params["starting_after"] = starting_after
    result = stripe.Charge.list(**params)
    for ch in result.data:
        charges.append({
            "id": ch.id,
            "amount": ch.amount / 100,
            "currency": ch.currency,
            "status": ch.status,
            "created": pd.to_datetime(ch.created, unit="s"),
            "customer": ch.customer,
            "description": ch.description,
        })
    has_more = result.has_more
    if result.data:
        starting_after = result.data[-1].id

df = pd.DataFrame(charges)
df.to_parquet(output_path, index=False)
print(f"Wrote {len(df)} charges to {output_path}")
`;
